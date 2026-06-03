package main

import "mailboxapi/internal/mailboxmodel"

func prepareMailboxProjection(mailbox *mailboxmodel.Record) {
	defaultMailboxProviderRegistry().PrepareProjection(mailbox)
}
