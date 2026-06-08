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

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	browserautomationv1 "mailboxapi/internal/contracts/browserautomationv1"
	"mailboxapi/internal/eventcatalog"
	"mailboxapi/internal/grpcclient"
	"mailboxapi/internal/grpchealth"
	"mailboxapi/internal/redisx"

	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxapp"
	"mailboxapi/internal/mailboxpg"
	"mailboxapi/pb"
)

func main() {
	cfg := loadConfig()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	browserConn, err := grpcclient.NewRequiredInsecure("browser automation", cfg.browserAutomationAddr)
	if err != nil {
		log.Fatalf("failed to connect browser automation: %s", safeMailboxError(err))
	}
	defer browserConn.Close()

	coordinationClient, err := newOptionalRedisClient(ctx, cfg.coordinationRedisURL, "coordination")
	if err != nil {
		log.Fatalf("failed to initialize mailbox coordination redis client: %s", safeMailboxError(err))
	}
	if coordinationClient != nil {
		defer func() { _ = coordinationClient.Close() }()
	}
	recentEmailClient, err := newOptionalRedisClient(ctx, cfg.recentEmailRedisURL, "recent email")
	if err != nil {
		log.Fatalf("failed to initialize mailbox recent email redis client: %s", safeMailboxError(err))
	}
	if recentEmailClient != nil {
		defer func() { _ = recentEmailClient.Close() }()
	}

	recentCache := newRecentEmailCache(recentEmailClient, cfg.recentEmailCachePrefix, cfg.recentEmailCacheTTL, cfg.recentEmailCacheMax)
	secretStore := newMailboxSecretStore(recentEmailClient, cfg.recentEmailCachePrefix+":secrets", cfg.recentEmailCacheTTL)
	mailboxRepo, err := mailboxpg.OpenRepository(ctx, cfg.pgDSN, cfg.providers.registry, mailboxPlatformEventOutboxTable)
	if err != nil {
		log.Fatalf("failed to initialize mailbox repository: %s", safeMailboxError(err))
	}
	defer mailboxRepo.Close()
	inboxService := inboxapp.NewService(inboxapp.Config{
		Repository:  mailboxRepo,
		Providers:   cfg.providers.registry,
		Recent:      recentCache,
		Secrets:     secretStore,
		SecretTTL:   cfg.recentEmailCacheTTL,
		OutboxTable: mailboxPlatformEventOutboxTable,
		EventSource: mailboxPlatformEventSource,
		Logf:        logWarning,
	})
	var inboxLock *redisx.BestEffortLocker
	if coordinationClient != nil {
		inboxLock = redisx.NewBestEffortLocker(coordinationClient, cfg.inboxLockPrefix, cfg.inboxLockTTL, cfg.inboxLockRetry)
	}
	platformEventBus, closePlatformEventBus, err := newPlatformEventBus(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize platform event bus: %s", safeMailboxError(err))
	}
	defer closePlatformEventBus()
	hotBus, closeHotStream, err := newMailboxHotStreamBus(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize mailbox hotstream: %s", safeMailboxError(err))
	}
	if closeHotStream != nil {
		defer closeHotStream()
	}
	hotEvents := newMailboxHotStream(hotBus)
	browserClient := browserautomationv1.NewBrowserAutomationServiceClient(browserConn)
	inboxSources := newMailboxInboxSourceRegistryForProviders(cfg.providers, mailboxInboxSourceDependencies{mailboxes: mailboxRepo})
	mailWatcher := NewMailWatcher(inboxService, mailboxRepo, inboxSources, hotEvents)

	operations, err := newOperationStore(cfg.pgDSN)
	if err != nil {
		log.Fatalf("failed to initialize mailbox operation store: %s", safeMailboxError(err))
	}
	var workDispatcher *mailboxWorkDispatcher
	if platformEventBus != nil {
		workDispatcher = newMailboxWorkDispatcher(operations.db, "mailbox-api")
	} else {
		logInfo("mailbox MQ dispatcher is disabled; mailbox operations run in local worker goroutines")
	}
	emailBackend := &EmailService{mailboxRepo: mailboxRepo, mailboxes: mailboxapp.NewService(mailboxRepo), inbox: inboxService, watcher: mailWatcher, providers: cfg.providers, inboxLock: inboxLock, work: workDispatcher}

	activities := newMailboxActivitiesForProviders(cfg.providers, mailboxProviderActionDependencies{browserClient: browserClient}, emailBackend, mailboxRepo, operations, hotEvents)

	errCh := make(chan error, 3)
	group, groupCtx := errgroup.WithContext(ctx)
	if platformEventBus != nil {
		platformEmailEvents := newMailboxPlatformEvents(platformEventBus)
		pollConsumer, err := platformEventBus.PullWorkerForDefinition(cfg.eventStreamName, eventcatalog.MailboxEmailPollRequested, 10, 60*time.Second)
		if err != nil {
			log.Fatalf("failed to initialize mailbox email poll worker: %s", safeMailboxError(err))
		}
		fetchConsumer, err := platformEventBus.PullWorkerForDefinition(cfg.eventStreamName, mailboxInboxFetchRequested, 5, 5*time.Minute)
		if err != nil {
			log.Fatalf("failed to initialize mailbox inbox fetch worker: %s", safeMailboxError(err))
		}
		registrationConsumer, err := platformEventBus.PullWorkerForDefinition(cfg.eventStreamName, mailboxRegistrationRequested, 2, 5*time.Minute)
		if err != nil {
			log.Fatalf("failed to initialize mailbox registration worker: %s", safeMailboxError(err))
		}
		oauthConsumer, err := platformEventBus.PullWorkerForDefinition(cfg.eventStreamName, mailboxOAuthRequested, 2, 5*time.Minute)
		if err != nil {
			log.Fatalf("failed to initialize mailbox OAuth worker: %s", safeMailboxError(err))
		}
		group.Go(func() error {
			return mailboxRepo.RunOutboxWorker(groupCtx, mailboxPlatformEventOutboxTable, platformEmailEvents, logWarning)
		})
		group.Go(func() error { return runMailboxEmailPollWorker(groupCtx, pollConsumer, emailBackend) })
		group.Go(func() error { return runMailboxInboxFetchWorker(groupCtx, fetchConsumer, emailBackend, operations) })
		group.Go(func() error {
			return runMailboxRegistrationWorker(groupCtx, registrationConsumer, operations, activities)
		})
		group.Go(func() error { return runMailboxOAuthWorker(groupCtx, oauthConsumer, operations, activities) })
	}
	startWebhookServer(groupCtx, cfg.webhookHTTPAddr, cfg.webhook, cfg.providers, inboxService, mailWatcher, inboxLock, errCh)

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

	dashboardConn, err := grpcclient.NewInsecure(grpcclient.SelfTarget(cfg.listenAddr))
	if err != nil {
		log.Fatalf("connect mailbox dashboard API: %s", safeMailboxError(err))
	}
	defer dashboardConn.Close()
	startDashboardHTTP(groupCtx, cfg.dashboardHTTPAddr, cfg.dashboardStaticDir, cfg.dashboard, pb.NewMailboxServiceClient(dashboardConn), hotBus, errCh)

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
