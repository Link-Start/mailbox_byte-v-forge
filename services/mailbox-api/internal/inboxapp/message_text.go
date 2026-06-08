package inboxapp

import (
	"html"
	"regexp"
	"strings"
	"time"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/hashx"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func UniqueEmails(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		trimmed := emailx.Normalize(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func CompactMessageText(value string, limit int) string {
	text := htmlTagPattern.ReplaceAllString(html.UnescapeString(value), " ")
	text = strings.Join(strings.Fields(strings.ReplaceAll(text, "\u00a0", " ")), " ")
	if limit > 0 && len(text) > limit {
		runes := []rune(text)
		if len(runes) > limit {
			return string(runes[:limit])
		}
	}
	return text
}

func MessageLimitValue(limit int32, fallback int) int {
	n := int(limit)
	if n <= 0 {
		n = fallback
	}
	if n <= 0 {
		n = 25
	}
	if n > 100 {
		n = 100
	}
	return n
}

func InboxReceivedAfter(watermarkNs int64, overlapSeconds int) int64 {
	if watermarkNs <= 0 {
		return 0
	}
	after := watermarkNs - int64(overlapSeconds)*int64(time.Second)
	if after < 0 {
		return 0
	}
	return after
}

func MessageMailboxEmails(accountEmail string, recipients []string) []string {
	items := []string{emailx.Normalize(accountEmail)}
	for _, recipient := range recipients {
		if email := emailx.Normalize(recipient); email != "" {
			items = append(items, email)
		}
	}
	return UniqueEmails(items)
}

func StableMessageKey(provider string, mailboxEmail string, value string) string {
	return hashx.StableParts(NormalizeProviderKey(provider), emailx.Normalize(mailboxEmail), strings.TrimSpace(value))
}

func NormalizeProviderKey(provider string) string {
	return strings.Trim(strings.ToLower(strings.TrimSpace(provider)), ".")
}
