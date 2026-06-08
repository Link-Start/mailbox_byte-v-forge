package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/dns"
	"github.com/cloudflare/cloudflare-go/v7/email_routing"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/cloudflare/cloudflare-go/v7/zones"
	"mailboxapi/internal/stringx"

	"mailboxapi/pb"
)

func fetchCloudflareEmailDomains(ctx context.Context, httpClient *http.Client, token string, baseURL string, cfg *pb.CloudflareEmailConfig) ([]string, error) {
	if cfg == nil {
		cfg = &pb.CloudflareEmailConfig{}
	}
	api := newCloudflareEmailAPI(httpClient, token, strings.TrimRight(stringx.FirstNonEmpty(baseURL, cfg.GetApiBaseUrl()), "/"))
	domains := newCloudflareEmailDomainCollector()
	zones, err := api.cloudflareEmailZones(ctx, cfg.GetZones())
	if err != nil {
		return nil, err
	}
	for _, zone := range zones {
		if !optionalBoolEnabled(zone.Enabled) || !optionalBoolEnabled(zone.CatchAllEnabled) {
			continue
		}
		zoneID, zoneName, err := api.resolveZone(ctx, zone)
		if err != nil {
			logWarning("skip Cloudflare email zone %s: %v", cloudflareZoneLabel(zone), err)
			continue
		}
		catchAll, err := api.emailRoutingCatchAll(ctx, zoneID)
		if err != nil {
			logWarning("skip Cloudflare email zone %s catch-all: %v", stringx.FirstNonEmpty(zoneName, cloudflareZoneLabel(zone)), err)
			continue
		}
		if !cloudflareCatchAllWorkerEnabled(catchAll, strings.TrimSpace(zone.GetWorkerName())) {
			continue
		}
		mxDomains, err := api.emailRoutingMXDomains(ctx, zoneID)
		if err != nil {
			logWarning("skip Cloudflare email zone %s MX: %v", stringx.FirstNonEmpty(zoneName, cloudflareZoneLabel(zone)), err)
			continue
		}
		if !domains.addAll(mxDomains) {
			domains.add(zoneName)
		}
	}
	return domains.values(), nil
}

type cloudflareEmailAPI struct {
	client *cloudflare.Client
}

func newCloudflareEmailAPI(httpClient *http.Client, token string, baseURL string) *cloudflareEmailAPI {
	if baseURL == "" {
		baseURL = defaultCloudflareAPIBaseURL
	}
	return &cloudflareEmailAPI{client: cloudflare.NewClient(
		option.WithAPIToken(token),
		option.WithHTTPClient(httpClient),
		option.WithBaseURL(baseURL),
		option.WithMaxRetries(0),
	)}
}

func (api *cloudflareEmailAPI) resolveZone(ctx context.Context, zone *pb.CloudflareEmailZone) (string, string, error) {
	zoneID := strings.TrimSpace(zone.GetZoneId())
	zoneName := normalizeCloudflareDomain(zone.GetZoneName())
	if zoneID != "" {
		return zoneID, zoneName, nil
	}
	if zoneName == "" {
		return "", "", fmt.Errorf("Cloudflare email zone requires zone_id or zone_name")
	}
	iter := api.client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{
		Name:    cloudflare.F(zoneName),
		PerPage: cloudflare.F(float64(1)),
	})
	if iter.Next() {
		found := iter.Current()
		return found.ID, normalizeCloudflareDomain(found.Name), nil
	}
	if err := iter.Err(); err != nil {
		return "", "", err
	}
	return "", "", fmt.Errorf("Cloudflare zone not found: %s", zoneName)
}

func (api *cloudflareEmailAPI) emailRoutingCatchAll(ctx context.Context, zoneID string) (*email_routing.RuleCatchAllGetResponse, error) {
	return api.client.EmailRouting.Rules.CatchAlls.Get(ctx, email_routing.RuleCatchAllGetParams{ZoneID: cloudflare.F(zoneID)})
}

func (api *cloudflareEmailAPI) emailRoutingMXDomains(ctx context.Context, zoneID string) ([]string, error) {
	iter := api.client.DNS.Records.ListAutoPaging(ctx, dns.RecordListParams{
		ZoneID:  cloudflare.F(zoneID),
		Type:    cloudflare.F(dns.RecordListParamsTypeMX),
		PerPage: cloudflare.F(float64(100)),
	})
	out := []string{}
	for iter.Next() {
		record := iter.Current()
		content := strings.ToLower(strings.TrimSpace(record.Content))
		if !strings.HasSuffix(content, ".mx.cloudflare.net") {
			continue
		}
		out = append(out, record.Name)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
