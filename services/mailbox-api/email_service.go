package main

import (
	"mailboxapi/internal/redisx"

	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxapp"
)

type EmailService struct {
	mailboxRepo     mailboxRepository
	mailboxes       *mailboxapp.Service
	inbox           *inboxapp.Service
	watcher         *MailWatcher
	providers       mailboxProviderRuntimeConfig
	inboxLock       *redisx.BestEffortLocker
	work            *mailboxWorkDispatcher
	cloudflareRelay *cloudflareEmailRelayClient
}
