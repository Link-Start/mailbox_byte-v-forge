package main

import (
	"log"
	"regexp"
	"strings"
)

const (
	defaultListenAddr          = ":50051"
	defaultWebhookTokenHeader  = "X-Webhook-Token"
	defaultPollIntervalSeconds = 5
	defaultMessageLimit        = 25
	defaultHTTPTimeoutSeconds  = 20
	defaultInboxOverlapSeconds = 120
	defaultWebhookMaxMailboxes = 100
	defaultWebhookTimeout      = 60
	defaultOutlookMaxMessages  = 100
	defaultCloudflareMaxDomain = 500
)

const (
	emailProviderOutlook    = "outlook"
	emailProviderCloudflare = "cloudflare"
)

var (
	emailPattern   = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
	htmlTagPattern = regexp.MustCompile(`<[^>]+>`)
)

func logInfo(format string, args ...any) {
	log.Printf("[MAIL] "+format, safeMailboxLogArgs(args...)...)
}

func logWarning(format string, args ...any) {
	log.Printf("[MAIL] WARNING "+format, safeMailboxLogArgs(args...)...)
}

func normalizeScope(value string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(value, ",", " ")), " ")
}
