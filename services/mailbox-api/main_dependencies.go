package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
	browserautomationv1 "mailboxapi/internal/contracts/browserautomationv1"
	"mailboxapi/internal/grpcclient"
	"mailboxapi/internal/mailboxmem"
	"mailboxapi/internal/mailboxpg"
)

func newBrowserAutomationClient(addr string) (browserautomationv1.BrowserAutomationServiceClient, func(), error) {
	if strings.TrimSpace(addr) == "" {
		logInfo("MAILBOX_BROWSER_AUTOMATION_ADDR is not configured; Outlook registration/OAuth browser actions are disabled")
		return nil, func() {}, nil
	}
	conn, err := grpcclient.NewRequiredInsecure("browser automation", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("connect browser automation: %w", err)
	}
	return browserautomationv1.NewBrowserAutomationServiceClient(conn), func() { _ = conn.Close() }, nil
}

func newMailboxRedisClients(ctx context.Context, cfg config) (*redis.Client, *redis.Client, func(), error) {
	coordination, err := newOptionalRedisClient(ctx, cfg.coordinationRedisURL, "coordination")
	if err != nil {
		return nil, nil, nil, err
	}
	recentEmail, err := newOptionalRedisClient(ctx, cfg.recentEmailRedisURL, "recent email")
	if err != nil {
		closeRedisClient(coordination)
		return nil, nil, nil, err
	}
	return coordination, recentEmail, func() {
		closeRedisClient(coordination)
		closeRedisClient(recentEmail)
	}, nil
}

func closeRedisClient(client *redis.Client) {
	if client != nil {
		_ = client.Close()
	}
}

func newMailboxRepository(ctx context.Context, cfg config) (mailboxRepository, error) {
	if cfg.pgDSN == "" {
		repository, err := mailboxmem.NewRepository(cfg.providers.registry)
		if err != nil {
			return nil, err
		}
		logInfo("mailbox postgres is disabled; mailbox data uses non-persistent in-process storage")
		return repository, nil
	}
	return mailboxpg.OpenRepository(ctx, cfg.pgDSN, cfg.providers.registry, mailboxEventOutboxTable)
}

func newMailboxOperationStore(ctx context.Context, cfg config) (operationStore, *pgOperationStore, error) {
	if cfg.pgDSN == "" {
		return newMemoryOperationStore(), nil, nil
	}
	store, err := newPgOperationStore(ctx, cfg.pgDSN)
	if err != nil {
		return nil, nil, err
	}
	return store, store, nil
}
