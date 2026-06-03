package main

import (
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/envx"
)

type outlookRuntimeConfig struct {
	registration outlookRegistrationConfig
	oauth        outlookOAuthConfig
	watcher      outlookWatcherConfig
}

type outlookOAuthConfig struct {
	clientID    string
	scope       string
	tokenURL    string
	httpTimeout time.Duration
}

type outlookWatcherConfig struct {
	messageLimit int
	pollInterval int
	inboxOverlap int
	httpTimeout  time.Duration
	oauth        outlookOAuthConfig
}

func loadOutlookRuntimeConfig() outlookRuntimeConfig {
	oauth := loadOutlookOAuthConfig()
	return outlookRuntimeConfig{
		registration: loadOutlookRegistrationConfig(),
		oauth:        oauth,
		watcher:      loadOutlookWatcherConfig(oauth),
	}
}

func loadOutlookOAuthConfig() outlookOAuthConfig {
	scope := normalizeScope(envx.StringDefault("OUTLOOK_OAUTH_SCOPE", outlookOAuthMailReadScope))
	if scope == "" {
		scope = outlookOAuthMailReadScope
	}
	return outlookOAuthConfig{
		clientID:    envx.StringDefault("OUTLOOK_OAUTH_CLIENT_ID", defaultOutlookOAuthClientID),
		scope:       scope,
		tokenURL:    envx.StringDefault("OUTLOOK_OAUTH_TOKEN_URL", defaultOutlookOAuthTokenURL),
		httpTimeout: positiveSeconds("OUTLOOK_HTTP_TIMEOUT_SECONDS", defaultHTTPTimeoutSeconds),
	}
}

func loadOutlookWatcherConfig(oauth outlookOAuthConfig) outlookWatcherConfig {
	messageLimit := envx.Int("OUTLOOK_MESSAGE_LIMIT", defaultMessageLimit)
	if messageLimit < 1 {
		messageLimit = 1
	}
	if messageLimit > 100 {
		messageLimit = 100
	}
	pollInterval := envx.Int("OUTLOOK_POLL_INTERVAL_SECONDS", defaultPollIntervalSeconds)
	if pollInterval < 1 {
		pollInterval = 1
	}
	inboxOverlap := envx.Int("OUTLOOK_INBOX_OVERLAP_SECONDS", defaultInboxOverlapSeconds)
	if inboxOverlap < 0 {
		inboxOverlap = 0
	}
	return outlookWatcherConfig{
		messageLimit: messageLimit,
		pollInterval: pollInterval,
		inboxOverlap: inboxOverlap,
		httpTimeout:  oauth.httpTimeout,
		oauth:        oauth,
	}
}

func (c outlookOAuthConfig) httpClient() *http.Client {
	return &http.Client{Timeout: c.httpTimeout}
}

func positiveSeconds(name string, fallback int) time.Duration {
	value := envx.Int(name, fallback)
	if value <= 0 {
		value = fallback
	}
	return time.Duration(value) * time.Second
}
