package main

import (
	"github.com/redis/go-redis/v9"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/natseventbus"
)

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
