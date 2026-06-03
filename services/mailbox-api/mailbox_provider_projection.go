package main

import "mailboxapi/internal/mailboxmodel"

func (c mailboxProviderRuntimeConfig) prepareProjection(mailbox *mailboxmodel.Record) {
	if c.registry != nil {
		c.registry.PrepareProjection(mailbox)
	}
}
