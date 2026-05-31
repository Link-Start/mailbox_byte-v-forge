package main

import "github.com/byte-v-forge/common-lib/redisx"

type EmailService struct {
	store     *MailboxStore
	watcher   *MailWatcher
	providers mailboxProviderRuntimeConfig
	inboxLock *redisx.BestEffortLocker
	work      *mailboxWorkDispatcher
}
