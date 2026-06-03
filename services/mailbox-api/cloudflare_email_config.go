package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/envx"
	"github.com/byte-v-forge/common-lib/protojsonx"

	"mailboxapi/pb"
)

const defaultCloudflareAPIBaseURL = "https://api.cloudflare.com/client/v4"

type cloudflareProviderConfig struct {
	apiToken             string
	apiBaseURL           string
	apiTimeout           time.Duration
	emailConfigFile      string
	maxMessagesPerDomain int
}

func loadCloudflareProviderConfig() cloudflareProviderConfig {
	return cloudflareProviderConfig{
		apiToken:             envx.String("MAILBOX_CLOUDFLARE_API_TOKEN"),
		apiBaseURL:           envx.StringDefault("MAILBOX_CLOUDFLARE_API_BASE_URL", ""),
		apiTimeout:           positiveSeconds("MAILBOX_CLOUDFLARE_API_TIMEOUT_SECONDS", defaultHTTPTimeoutSeconds),
		emailConfigFile:      envx.String("MAILBOX_CLOUDFLARE_EMAIL_CONFIG_FILE"),
		maxMessagesPerDomain: envx.Int("MAILBOX_CLOUDFLARE_MAX_MESSAGES_PER_DOMAIN", defaultCloudflareMaxDomain),
	}
}

func (c cloudflareProviderConfig) loadEmailDomains() []string {
	cfg := loadCloudflareEmailConfig(c.emailConfigFile)
	if c.apiToken == "" {
		if cfg != nil && len(cfg.GetZones()) > 0 {
			logWarning("MAILBOX_CLOUDFLARE_API_TOKEN is required to load Cloudflare email domains")
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.apiTimeout)
	defer cancel()
	domains, err := fetchCloudflareEmailDomains(ctx, &http.Client{Timeout: c.apiTimeout}, c.apiToken, c.apiBaseURL, cfg)
	if err != nil {
		logWarning("fetch Cloudflare email config: %v", err)
		return nil
	}
	if len(domains) == 0 {
		logWarning("Cloudflare email API returned no mailbox domains")
		return nil
	}
	logInfo("loaded Cloudflare email domains from API count=%d", len(domains))
	return domains
}

func loadCloudflareEmailConfig(path string) *pb.CloudflareEmailConfig {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		logWarning("read Cloudflare email config: %v", err)
		return nil
	}
	var cfg pb.CloudflareEmailConfig
	if err := protojsonx.Unmarshal(raw, &cfg); err != nil {
		logWarning("decode Cloudflare email config: %v", err)
		return nil
	}
	return &cfg
}
