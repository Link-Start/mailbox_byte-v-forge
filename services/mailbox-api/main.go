package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/byte-v-forge/common-lib/eventcatalog"
	browserautomationv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/browserautomation/v1"
	"github.com/byte-v-forge/common-lib/grpcclient"
	"github.com/byte-v-forge/common-lib/grpchealth"
	"github.com/byte-v-forge/common-lib/redisx"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"mailboxapi/pb"
)

func main() {
	cfg := loadConfig()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	browserConn, err := newGRPCClient("browser automation", cfg.browserAutomationAddr)
	if err != nil {
		log.Fatalf("failed to connect browser automation: %s", safeMailboxError(err))
	}
	defer browserConn.Close()

	coordinationClient, closeCoordinationClient, err := newRequiredRedisClient(ctx, cfg.coordinationRedisURL, "MAILBOX_COORDINATION_REDIS_URL is required for mailbox coordination")
	if err != nil {
		log.Fatalf("failed to initialize mailbox coordination redis client: %s", safeMailboxError(err))
	}
	if closeCoordinationClient != nil {
		defer func() { _ = closeCoordinationClient() }()
	}
	recentEmailClient, closeRecentEmailClient, err := newRequiredRedisClient(ctx, cfg.recentEmailRedisURL, "MAILBOX_RECENT_EMAIL_REDIS_URL is required for mailbox recent email cache")
	if err != nil {
		log.Fatalf("failed to initialize mailbox recent email redis client: %s", safeMailboxError(err))
	}
	if closeRecentEmailClient != nil {
		defer func() { _ = closeRecentEmailClient() }()
	}

	recentCache := newRecentEmailCache(recentEmailClient, cfg.recentEmailCachePrefix, cfg.recentEmailCacheTTL, cfg.recentEmailCacheMax)
	mailboxStore, err := NewMailboxStore(ctx, cfg.pgDSN, recentCache)
	if err != nil {
		log.Fatalf("failed to initialize mailbox store: %s", safeMailboxError(err))
	}
	defer mailboxStore.Close()
	inboxLock := redisx.NewBestEffortLocker(coordinationClient, cfg.inboxLockPrefix, cfg.inboxLockTTL, cfg.inboxLockRetry)
	platformEventBus, closePlatformEventBus, err := newPlatformEventBus(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize platform event bus: %s", safeMailboxError(err))
	}
	if closePlatformEventBus != nil {
		defer closePlatformEventBus()
	}
	if platformEventBus == nil {
		log.Fatal("PLATFORM_NATS_URL is required for mailbox event workers")
	}
	hotBus, closeHotStream, err := newMailboxHotStreamBus(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize mailbox hotstream: %s", safeMailboxError(err))
	}
	if closeHotStream != nil {
		defer closeHotStream()
	}
	hotEvents := newMailboxHotStream(hotBus)
	platformEmailEvents := newMailboxPlatformEvents(platformEventBus)
	mailWatcher := NewMailWatcher(mailboxStore, hotEvents)

	operations, err := newOperationStore(cfg.pgDSN)
	if err != nil {
		log.Fatalf("failed to initialize mailbox operation store: %s", safeMailboxError(err))
	}
	workDispatcher := newMailboxWorkDispatcher(operations.db, "mailbox-api")
	emailBackend := &EmailService{store: mailboxStore, watcher: mailWatcher, providers: cfg.providers, inboxLock: inboxLock, work: workDispatcher}

	pollConsumer, err := platformEventBus.PullWorkerConsumer(cfg.eventStreamName, eventcatalog.MailboxEmailPollRequested.Subject, eventcatalog.MailboxEmailPollRequested.ConsumerDurable, 10, 60*time.Second)
	if err != nil {
		log.Fatalf("failed to initialize mailbox email poll worker: %s", safeMailboxError(err))
	}

	fetchConsumer, err := platformEventBus.PullWorkerConsumer(cfg.eventStreamName, eventcatalog.MailboxInboxFetchRequested.Subject, eventcatalog.MailboxInboxFetchRequested.ConsumerDurable, 5, 5*time.Minute)
	if err != nil {
		log.Fatalf("failed to initialize mailbox inbox fetch worker: %s", safeMailboxError(err))
	}

	activities := newMailboxActivitiesForProviders(cfg.providers, browserautomationv1.NewBrowserAutomationServiceClient(browserConn), emailBackend, operations, hotEvents)

	registrationConsumer, err := platformEventBus.PullWorkerConsumer(cfg.eventStreamName, eventcatalog.MailboxRegistrationRequested.Subject, eventcatalog.MailboxRegistrationRequested.ConsumerDurable, 2, 5*time.Minute)
	if err != nil {
		log.Fatalf("failed to initialize mailbox registration worker: %s", safeMailboxError(err))
	}

	oauthConsumer, err := platformEventBus.PullWorkerConsumer(cfg.eventStreamName, eventcatalog.MailboxOAuthRequested.Subject, eventcatalog.MailboxOAuthRequested.ConsumerDurable, 2, 5*time.Minute)
	if err != nil {
		log.Fatalf("failed to initialize mailbox OAuth worker: %s", safeMailboxError(err))
	}

	errCh := make(chan error, 3)
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error { return runMailboxPlatformEventOutboxWorker(groupCtx, mailboxStore, platformEmailEvents) })
	group.Go(func() error { return runMailboxEmailPollWorker(groupCtx, pollConsumer, emailBackend) })
	group.Go(func() error { return runMailboxInboxFetchWorker(groupCtx, fetchConsumer, emailBackend, operations) })
	group.Go(func() error {
		return runMailboxRegistrationWorker(groupCtx, registrationConsumer, operations, activities)
	})
	group.Go(func() error { return runMailboxOAuthWorker(groupCtx, oauthConsumer, operations, activities) })
	startWebhookServer(groupCtx, cfg.webhookHTTPAddr, mailboxStore, mailWatcher, inboxLock, errCh)

	listener, err := net.Listen("tcp", cfg.listenAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", cfg.listenAddr, safeMailboxError(err))
	}

	mailboxServer := &server{
		emailBackend: emailBackend,
		operations:   operations,
		providers:    cfg.providers,
		hot:          hotEvents,
		work:         workDispatcher,
		activities:   activities,
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMailboxServiceServer(grpcServer, mailboxServer)
	grpchealth.RegisterServing(grpcServer)

	dashboardConn, err := grpcclient.NewInsecure(selfTarget(cfg.listenAddr))
	if err != nil {
		log.Fatalf("connect mailbox dashboard API: %s", safeMailboxError(err))
	}
	defer dashboardConn.Close()
	startDashboardHTTP(groupCtx, cfg.dashboardHTTPAddr, cfg.dashboardStaticDir, pb.NewMailboxServiceClient(dashboardConn), hotBus, errCh)

	go func() {
		<-groupCtx.Done()
		grpcServer.GracefulStop()
	}()

	log.Printf("mailbox API listening on %s", cfg.listenAddr)
	group.Go(func() error {
		if err := grpcServer.Serve(listener); err != nil {
			return fmt.Errorf("mailbox API failed: %w", err)
		}
		return nil
	})
	group.Go(func() error {
		select {
		case <-groupCtx.Done():
			return nil
		case err := <-errCh:
			return err
		}
	})
	if err := group.Wait(); err != nil {
		stop()
		log.Fatal(safeMailboxError(err))
	}
}
