package main

import (
	"github.com/byte-v-forge/common-lib/redisx"

	"mailboxapi/internal/mailboxapp"
)

type EmailService struct {
	store     *MailboxStore
	mailboxes *mailboxapp.Service
	watcher   *MailWatcher
	providers mailboxProviderRuntimeConfig
	inboxLock *redisx.BestEffortLocker
	work      *mailboxWorkDispatcher
}
