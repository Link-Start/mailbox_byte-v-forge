package main

import (
	"github.com/byte-v-forge/common-lib/redisx"

	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxapp"
	"mailboxapi/internal/mailboxpg"
)

type EmailService struct {
	mailboxRepo *mailboxpg.Repository
	mailboxes   *mailboxapp.Service
	inbox       *inboxapp.Service
	watcher     *MailWatcher
	providers   mailboxProviderRuntimeConfig
	inboxLock   *redisx.BestEffortLocker
	work        *mailboxWorkDispatcher
}
