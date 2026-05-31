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

func loadCloudflareEmailDomains() []string {
	cfg := loadCloudflareEmailConfig()
	token := strings.TrimSpace(os.Getenv("MAILBOX_CLOUDFLARE_API_TOKEN"))
	if token == "" {
		if cfg != nil && len(cfg.GetZones()) > 0 {
			logWarning("MAILBOX_CLOUDFLARE_API_TOKEN is required to load Cloudflare email domains")
		}
		return nil
	}
	timeout := time.Duration(envx.Int("MAILBOX_CLOUDFLARE_API_TIMEOUT_SECONDS", defaultHTTPTimeoutSeconds)) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	domains, err := fetchCloudflareEmailDomains(ctx, &http.Client{Timeout: timeout}, token, cfg)
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

func loadCloudflareEmailConfig() *pb.CloudflareEmailConfig {
	path := strings.TrimSpace(os.Getenv("MAILBOX_CLOUDFLARE_EMAIL_CONFIG_FILE"))
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
