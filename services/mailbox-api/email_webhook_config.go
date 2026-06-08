package main

import (
	"time"

	"mailboxapi/internal/envx"
)

type emailWebhookConfig struct {
	token                    string
	outlookFetchTimeout      time.Duration
	outlookRefreshMaxMailbox int
}

func loadEmailWebhookConfig() emailWebhookConfig {
	return emailWebhookConfig{
		token:                    envx.String("MAILBOX_WEBHOOK_TOKEN"),
		outlookFetchTimeout:      positiveSeconds("OUTLOOK_WEBHOOK_FETCH_TIMEOUT_SECONDS", defaultWebhookTimeout),
		outlookRefreshMaxMailbox: positiveInt("OUTLOOK_WEBHOOK_MAX_MAILBOXES", defaultWebhookMaxMailboxes),
	}
}

func positiveInt(name string, fallback int) int {
	value := envx.Int(name, fallback)
	if value <= 0 {
		value = fallback
	}
	return value
}
