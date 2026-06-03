package main

import (
	"mailboxapi/internal/mailboxpg"
)

func domainForEmail(email string) string {
	return mailboxpg.DomainForEmail(email)
}
