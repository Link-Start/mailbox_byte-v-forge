package main

import (
	"github.com/byte-v-forge/common-lib/redisx"

	"mailboxapi/internal/mailboxapp"
	"mailboxapi/internal/mailboxpg"
)

type EmailService struct {
	store       *MailboxStore
	mailboxRepo *mailboxpg.Repository
	mailboxes   *mailboxapp.Service
	watcher     *MailWatcher
	providers   mailboxProviderRuntimeConfig
	inboxLock   *redisx.BestEffortLocker
	work        *mailboxWorkDispatcher
}
