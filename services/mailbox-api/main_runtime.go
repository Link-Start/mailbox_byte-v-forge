package main

import (
	"context"
	"fmt"

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
