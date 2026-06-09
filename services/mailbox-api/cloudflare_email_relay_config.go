package main

import (
	"strings"
	"time"

	"mailboxapi/internal/envx"
)

const (
	defaultCloudflareRelayPullTimeoutSeconds = 10
	defaultCloudflareRelayPullMaxEvents      = 50
)

type cloudflareRelayPullConfig struct {
	baseURL   string
	token     string
	timeout   time.Duration
	maxEvents int
}

func loadCloudflareRelayPullConfig() cloudflareRelayPullConfig {
	return cloudflareRelayPullConfig{
		baseURL:   strings.TrimRight(envx.String("MAILBOX_CLOUDFLARE_RELAY_PULL_URL"), "/"),
		token:     envx.String("MAILBOX_CLOUDFLARE_RELAY_PULL_TOKEN"),
		timeout:   positiveSeconds("MAILBOX_CLOUDFLARE_RELAY_PULL_TIMEOUT_SECONDS", defaultCloudflareRelayPullTimeoutSeconds),
		maxEvents: positiveInt("MAILBOX_CLOUDFLARE_RELAY_PULL_MAX_EVENTS", defaultCloudflareRelayPullMaxEvents),
	}
}

func (c cloudflareRelayPullConfig) enabled() bool {
	return strings.TrimSpace(c.baseURL) != "" && strings.TrimSpace(c.token) != ""
}
