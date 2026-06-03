package main

import (
	"time"

	"github.com/byte-v-forge/common-lib/envx"
)

type dashboardConfig struct {
	inboxTimeout time.Duration
}

func loadDashboardConfig() dashboardConfig {
	timeout := time.Duration(envx.Int("MAILBOX_INBOX_TIMEOUT_SECONDS", 180)) * time.Second
	if timeout < 30*time.Second {
		timeout = 30 * time.Second
	}
	return dashboardConfig{inboxTimeout: timeout}
}
