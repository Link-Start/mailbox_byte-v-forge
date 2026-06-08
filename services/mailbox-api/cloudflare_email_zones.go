package main

import (
	"context"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/zones"
	"mailboxapi/internal/stringx"

	"mailboxapi/pb"
)

type cloudflareEmailZoneCollector struct {
	zones []*pb.CloudflareEmailZone
	seen  map[string]struct{}
}

func newCloudflareEmailZoneCollector() *cloudflareEmailZoneCollector {
	return &cloudflareEmailZoneCollector{seen: map[string]struct{}{}}
}

func (c *cloudflareEmailZoneCollector) add(zone *pb.CloudflareEmailZone) {
	if zone == nil {
		return
	}
	key := stringx.FirstNonEmpty(strings.TrimSpace(zone.GetZoneId()), normalizeCloudflareDomain(zone.GetZoneName()))
	if key == "" {
		return
	}
	if _, ok := c.seen[key]; ok {
		return
	}
	c.seen[key] = struct{}{}
	c.zones = append(c.zones, zone)
}

func (c *cloudflareEmailZoneCollector) values() []*pb.CloudflareEmailZone {
	return c.zones
}

func (api *cloudflareEmailAPI) cloudflareEmailZones(ctx context.Context, configured []*pb.CloudflareEmailZone) ([]*pb.CloudflareEmailZone, error) {
	collector := newCloudflareEmailZoneCollector()
	for _, zone := range configured {
		collector.add(zone)
	}

	iter := api.client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{PerPage: cloudflare.F(float64(50))})
	for iter.Next() {
		zone := iter.Current()
		name := normalizeCloudflareDomain(zone.Name)
		if zone.ID == "" || name == "" {
			continue
		}
		collector.add(&pb.CloudflareEmailZone{ZoneId: zone.ID, ZoneName: name})
	}
	if err := iter.Err(); err != nil {
		if len(collector.values()) > 0 {
			logWarning("skip dynamic Cloudflare zone discovery: %v", err)
			return collector.values(), nil
		}
		return nil, err
	}
	return collector.values(), nil
}
