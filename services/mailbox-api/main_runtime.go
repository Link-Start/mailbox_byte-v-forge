package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"mailboxapi/internal/hotstream"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxapp"
	"mailboxapi/internal/natseventbus"
	"mailboxapi/internal/redisx"
)

type mailboxRuntime struct {
	repository     mailboxRepository
	providers      mailboxProviderRuntimeConfig
	eventBus       *natseventbus.Bus
	hotBus         hotstream.Bus
	hotEvents      *mailboxHotStream
	inbox          *inboxapp.Service
	watcher        *MailWatcher
	inboxLock      *redisx.BestEffortLocker
	operations     operationStore
	activities     *mailboxActivities
	emailBackend   *EmailService
	workDispatcher *mailboxWorkDispatcher
}

type mailboxRuntimeClosers []func()

func newMailboxRuntime(ctx context.Context, cfg config) (*mailboxRuntime, func(), error) {
	var closers mailboxRuntimeClosers
	committed := false
	defer func() {
		if !committed {
			closers.close()
		}
	}()

	browserClient, closeBrowserClient, err := newBrowserAutomationClient(cfg.browserAutomationAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("browser automation client: %w", err)
	}
	closers.add(closeBrowserClient)

	coordinationClient, recentEmailClient, closeRedisClients, err := newMailboxRedisClients(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("mailbox redis clients: %w", err)
	}
	closers.add(closeRedisClients)

	repository, err := newMailboxRepository(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("mailbox repository: %w", err)
	}
	closers.add(repository.Close)

	eventBus, closeEventBus, err := newMailboxEventBus(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("mailbox event bus: %w", err)
	}
	closers.add(closeEventBus)

	hotBus, closeHotStream, err := newMailboxHotStreamBus(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("mailbox hotstream: %w", err)
	}
	closers.add(closeHotStream)

	operations, pgOperations, err := newMailboxOperationStore(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("mailbox operation store: %w", err)
	}
	if pgOperations != nil {
		closers.add(pgOperations.Close)
	}

	hotEvents := newMailboxHotStream(hotBus)
	inbox := newMailboxInboxService(repository, recentEmailClient, cfg)
	inboxLock := newMailboxInboxLock(coordinationClient, cfg)
	watcher := NewMailWatcher(
		inbox,
		repository,
		newMailboxInboxSourceRegistryForProviders(cfg.providers, mailboxInboxSourceDependencies{mailboxes: repository}),
		hotEvents,
	)
	workDispatcher := newMailboxRuntimeWorkDispatcher(eventBus, pgOperations)
	emailBackend := &EmailService{
		mailboxRepo: repository,
		mailboxes:   mailboxapp.NewService(repository),
		inbox:       inbox,
		watcher:     watcher,
		providers:   cfg.providers,
		inboxLock:   inboxLock,
		work:        workDispatcher,
	}
	activities := newMailboxActivitiesForProviders(
		cfg.providers,
		mailboxProviderActionDependencies{browserClient: browserClient},
		emailBackend,
		repository,
		operations,
		hotEvents,
	)
	committed = true
	return &mailboxRuntime{
		repository:     repository,
		providers:      cfg.providers,
		eventBus:       eventBus,
		hotBus:         hotBus,
		hotEvents:      hotEvents,
		inbox:          inbox,
		watcher:        watcher,
		inboxLock:      inboxLock,
		operations:     operations,
		activities:     activities,
		emailBackend:   emailBackend,
		workDispatcher: workDispatcher,
	}, closers.close, nil
}

func newMailboxInboxService(repository mailboxRepository, recentEmailClient redis.Cmdable, cfg config) *inboxapp.Service {
	return inboxapp.NewService(inboxapp.Config{
		Repository:  repository,
		Providers:   cfg.providers.registry,
		Recent:      newRecentEmailCache(recentEmailClient, cfg.recentEmailCachePrefix, cfg.recentEmailCacheTTL, cfg.recentEmailCacheMax),
		Secrets:     newMailboxSecretStore(recentEmailClient, cfg.recentEmailCachePrefix+":secrets", cfg.recentEmailCacheTTL),
		SecretTTL:   cfg.recentEmailCacheTTL,
		OutboxTable: mailboxEventOutboxTable,
		EventSource: mailboxEventSource,
		Logf:        logWarning,
	})
}

func newMailboxRuntimeWorkDispatcher(eventBus *natseventbus.Bus, operations *pgOperationStore) *mailboxWorkDispatcher {
	if eventBus != nil && operations != nil {
		return newMailboxWorkDispatcher(operations.pool, "mailbox-api")
	}
	logInfo("mailbox MQ dispatcher is disabled; mailbox operations run in local worker goroutines")
	return nil
}

func (r *mailboxRuntime) serverRuntime() mailboxServerRuntime {
	return mailboxServerRuntime{
		emailBackend:    r.emailBackend,
		operations:      r.operations,
		activities:      r.activities,
		providers:       r.providers,
		hot:             r.hotEvents,
		work:            r.workDispatcher,
		inbox:           r.inbox,
		watcher:         r.watcher,
		inboxLock:       r.inboxLock,
		dashboardEvents: r.hotBus,
	}
}

func (c *mailboxRuntimeClosers) add(close func()) {
	if close != nil {
		*c = append(*c, close)
	}
}

func (c mailboxRuntimeClosers) close() {
	for i := len(c) - 1; i >= 0; i-- {
		c[i]()
	}
}
