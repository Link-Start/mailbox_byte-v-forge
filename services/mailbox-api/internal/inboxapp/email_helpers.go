package inboxapp

import "mailboxapi/internal/emailx"

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

func MessageMailboxEmails(accountEmail string, recipients []string) []string {
	items := []string{emailx.Normalize(accountEmail)}
	for _, recipient := range recipients {
		if email := emailx.Normalize(recipient); email != "" {
			items = append(items, email)
		}
	}
	return UniqueEmails(items)
}
