package main

import (
	"mailboxapi/internal/mailboxpg"
)

func normalizeEmailProvider(provider string) string {
	return normalizeMailboxProviderInput(provider)
}

func domainForEmail(email string) string {
	return mailboxpg.DomainForEmail(email)
}
