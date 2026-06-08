package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxapp"
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
	inboxLock := newMailboxInboxLock(coordinationClient, cfg)
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

	group, groupCtx := errgroup.WithContext(ctx)
	if err := startMailboxEventWorkers(groupCtx, group, cfg, mailboxEventBus, mailboxRepo, emailBackend, operations, activities); err != nil {
		log.Fatalf("failed to initialize mailbox event workers: %s", safeMailboxError(err))
	}
	if err := startMailboxServers(groupCtx, group, cfg, mailboxServerRuntime{
		emailBackend:    emailBackend,
		operations:      operations,
		activities:      activities,
		providers:       cfg.providers,
		hot:             hotEvents,
		work:            workDispatcher,
		inbox:           inboxService,
		watcher:         mailWatcher,
		inboxLock:       inboxLock,
		dashboardEvents: hotBus,
	}); err != nil {
		log.Fatalf("failed to initialize mailbox servers: %s", safeMailboxError(err))
	}
	if err := group.Wait(); err != nil {
		stop()
		log.Fatal(safeMailboxError(err))
	}
}
