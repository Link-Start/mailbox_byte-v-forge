package main

import (
	"log"
	"time"

	"github.com/byte-v-forge/common-lib/envx"
	"github.com/byte-v-forge/common-lib/natseventbus"
)

type config struct {
	listenAddr             string
	pgDSN                  string
	webhookHTTPAddr        string
	dashboardHTTPAddr      string
	dashboardStaticDir     string
	browserAutomationAddr  string
	coordinationRedisURL   string
	recentEmailRedisURL    string
	recentEmailCachePrefix string
	recentEmailCacheTTL    time.Duration
	recentEmailCacheMax    int
	platformNATSURL        string
	eventStreamName        string
	inboxLockPrefix        string
	inboxLockTTL           time.Duration
	inboxLockRetry         time.Duration
	providers              mailboxProviderRuntimeConfig
	outlook                outlookRuntimeConfig
	webhook                emailWebhookConfig
}

func loadConfig() config {
	return config{
		listenAddr:             envx.StringDefault("LISTEN_ADDR", ":50051"),
		pgDSN:                  requiredEnv("MAILBOX_PG_DSN"),
		webhookHTTPAddr:        envx.StringDefault("MAILBOX_WEBHOOK_HTTP_ADDR", ":8082"),
		dashboardHTTPAddr:      envx.StringDefault("MAILBOX_DASHBOARD_HTTP_ADDR", ":8080"),
		dashboardStaticDir:     envx.StringDefault("MAILBOX_DASHBOARD_STATIC_DIR", "/app/dashboard/mailbox"),
		browserAutomationAddr:  envx.StringDefault("BROWSER_AUTOMATION_ADDR", "browser-automation:50051"),
		coordinationRedisURL:   envx.StringDefault("MAILBOX_COORDINATION_REDIS_URL", ""),
		recentEmailRedisURL:    envx.StringDefault("MAILBOX_RECENT_EMAIL_REDIS_URL", ""),
		recentEmailCachePrefix: envx.StringDefault("MAILBOX_RECENT_EMAIL_CACHE_KEY_PREFIX", "byte-v-forge:mailbox:recent-email"),
		recentEmailCacheTTL:    envx.PositiveDurationSeconds("MAILBOX_RECENT_EMAIL_CACHE_TTL_SECONDS", time.Hour),
		recentEmailCacheMax:    envx.PositiveInt("MAILBOX_RECENT_EMAIL_CACHE_MAX_MESSAGES", 20),
		platformNATSURL:        envx.StringDefault("PLATFORM_NATS_URL", ""),
		eventStreamName:        envx.StringDefault("PLATFORM_EVENT_STREAM_NAME", natseventbus.DefaultStream),
		inboxLockPrefix:        envx.StringDefault("MAILBOX_INBOX_LOCK_KEY_PREFIX", "byte-v-forge:mailbox:locks"),
		inboxLockTTL:           envx.PositiveDurationSeconds("MAILBOX_INBOX_LOCK_TTL_SECONDS", 10*time.Minute),
		inboxLockRetry:         envx.PositiveDurationSeconds("MAILBOX_INBOX_LOCK_RETRY_SECONDS", time.Second),
		providers:              loadMailboxProviderRuntimeConfig(loadMailboxProviderConfig()),
		outlook:                loadOutlookRuntimeConfig(),
		webhook:                loadEmailWebhookConfig(),
	}
}

func requiredEnv(name string) string {
	if value := envx.String(name); value != "" {
		return value
	}
	log.Fatalf("%s is required", name)
	return ""
}
