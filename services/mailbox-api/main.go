package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"mailboxapi/internal/grpcclient"
	"mailboxapi/internal/grpchealth"
	"mailboxapi/internal/redisx"

	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxapp"
	"mailboxapi/pb"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("failed to load mailbox config: %s", safeMailboxError(err))
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	browserClient, closeBrowserClient, err := newBrowserAutomationClient(cfg.browserAutomationAddr)
	if err != nil {
		log.Fatalf("failed to initialize browser automation client: %s", safeMailboxError(err))
	}
	defer closeBrowserClient()

	coordinationClient, recentEmailClient, closeRedisClients, err := newMailboxRedisClients(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize mailbox redis clients: %s", safeMailboxError(err))
	}
	defer closeRedisClients()

	recentCache := newRecentEmailCache(recentEmailClient, cfg.recentEmailCachePrefix, cfg.recentEmailCacheTTL, cfg.recentEmailCacheMax)
	secretStore := newMailboxSecretStore(recentEmailClient, cfg.recentEmailCachePrefix+":secrets", cfg.recentEmailCacheTTL)
	mailboxRepo, err := newMailboxRepository(ctx, cfg)
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
		OutboxTable: mailboxEventOutboxTable,
		EventSource: mailboxEventSource,
		Logf:        logWarning,
	})
	var inboxLock *redisx.BestEffortLocker
	if coordinationClient != nil {
		inboxLock = redisx.NewBestEffortLocker(coordinationClient, cfg.inboxLockPrefix, cfg.inboxLockTTL, cfg.inboxLockRetry)
	}
	mailboxEventBus, closeMailboxEventBus, err := newMailboxEventBus(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize mailbox event bus: %s", safeMailboxError(err))
	}
	defer closeMailboxEventBus()
	hotBus, closeHotStream, err := newMailboxHotStreamBus(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize mailbox hotstream: %s", safeMailboxError(err))
	}
	if closeHotStream != nil {
		defer closeHotStream()
	}
	hotEvents := newMailboxHotStream(hotBus)
	inboxSources := newMailboxInboxSourceRegistryForProviders(cfg.providers, mailboxInboxSourceDependencies{mailboxes: mailboxRepo})
	mailWatcher := NewMailWatcher(inboxService, mailboxRepo, inboxSources, hotEvents)

	operations, pgOperations, err := newMailboxOperationStore(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize mailbox operation store: %s", safeMailboxError(err))
	}
	if pgOperations != nil {
		defer pgOperations.Close()
	}
	var workDispatcher *mailboxWorkDispatcher
	if mailboxEventBus != nil && pgOperations != nil {
		workDispatcher = newMailboxWorkDispatcher(pgOperations.pool, "mailbox-api")
	} else {
		logInfo("mailbox MQ dispatcher is disabled; mailbox operations run in local worker goroutines")
	}
	emailBackend := &EmailService{mailboxRepo: mailboxRepo, mailboxes: mailboxapp.NewService(mailboxRepo), inbox: inboxService, watcher: mailWatcher, providers: cfg.providers, inboxLock: inboxLock, work: workDispatcher}

	activities := newMailboxActivitiesForProviders(cfg.providers, mailboxProviderActionDependencies{browserClient: browserClient}, emailBackend, mailboxRepo, operations, hotEvents)

	errCh := make(chan error, 3)
	group, groupCtx := errgroup.WithContext(ctx)
	if err := startMailboxEventWorkers(groupCtx, group, cfg, mailboxEventBus, mailboxRepo, emailBackend, operations, activities); err != nil {
		log.Fatalf("failed to initialize mailbox event workers: %s", safeMailboxError(err))
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
