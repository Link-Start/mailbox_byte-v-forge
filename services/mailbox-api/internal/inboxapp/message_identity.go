package inboxapp

import (
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/hashx"
)

func StableMessageKey(provider string, mailboxEmail string, value string) string {
	return hashx.StableParts(NormalizeProviderKey(provider), emailx.Normalize(mailboxEmail), strings.TrimSpace(value))
}

func NormalizeProviderKey(provider string) string {
	return strings.Trim(strings.ToLower(strings.TrimSpace(provider)), ".")
}
