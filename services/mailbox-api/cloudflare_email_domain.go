package main

import (
	"strings"

	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/cloudflare/cloudflare-go/v7/email_routing"

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

func appendCloudflareEmailDomain(out *[]string, seen map[string]struct{}, value string) bool {
	domain := normalizeCloudflareDomain(value)
	if domain == "" {
		return false
	}
	if _, ok := seen[domain]; ok {
		return false
	}
	seen[domain] = struct{}{}
	*out = append(*out, domain)
	return true
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
