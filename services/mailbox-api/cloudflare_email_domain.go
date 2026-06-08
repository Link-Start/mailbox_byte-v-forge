package main

import (
	"strings"

	"github.com/cloudflare/cloudflare-go/v7/email_routing"
	"mailboxapi/internal/stringx"

	"mailboxapi/pb"
)

func cloudflareCatchAllWorkerEnabled(rule *email_routing.RuleCatchAllGetResponse, workerName string) bool {
	if rule == nil || !bool(rule.Enabled) {
		return false
	}
	for _, action := range rule.Actions {
		if action.Type != email_routing.CatchAllActionTypeWorker && strings.TrimSpace(strings.ToLower(string(action.Type))) != "worker" {
			continue
		}
		if workerName == "" {
			return true
		}
		for _, value := range action.Value {
			if strings.EqualFold(strings.TrimSpace(value), workerName) {
				return true
			}
		}
	}
	return false
}

func optionalBoolEnabled(value *bool) bool {
	return value == nil || *value
}

type cloudflareEmailDomainCollector struct {
	domains []string
	seen    map[string]struct{}
}

func newCloudflareEmailDomainCollector() *cloudflareEmailDomainCollector {
	return &cloudflareEmailDomainCollector{seen: map[string]struct{}{}}
}

func (c *cloudflareEmailDomainCollector) add(value string) bool {
	domain := normalizeCloudflareDomain(value)
	if domain == "" {
		return false
	}
	if _, ok := c.seen[domain]; ok {
		return false
	}
	c.seen[domain] = struct{}{}
	c.domains = append(c.domains, domain)
	return true
}

func (c *cloudflareEmailDomainCollector) addAll(values []string) bool {
	added := false
	for _, value := range values {
		added = c.add(value) || added
	}
	return added
}

func (c *cloudflareEmailDomainCollector) values() []string {
	return c.domains
}

func normalizeCloudflareDomain(value string) string {
	return strings.Trim(strings.ToLower(strings.TrimSpace(value)), ".")
}

func cloudflareZoneLabel(zone *pb.CloudflareEmailZone) string {
	if zone == nil {
		return "unknown"
	}
	return stringx.FirstNonEmpty(normalizeCloudflareDomain(zone.GetZoneName()), strings.TrimSpace(zone.GetZoneId()), "unknown")
}
