package main

import "github.com/byte-v-forge/common-lib/envx"

type mailboxProviderConfig struct {
	outlookMaxMessages int
	cloudflare         cloudflareProviderConfig
}

func loadMailboxProviderConfig() mailboxProviderConfig {
	return mailboxProviderConfig{
		outlookMaxMessages: envx.Int("MAILBOX_OUTLOOK_MAX_MESSAGES_PER_MAILBOX", defaultOutlookMaxMessages),
		cloudflare:         loadCloudflareProviderConfig(),
	}
}
