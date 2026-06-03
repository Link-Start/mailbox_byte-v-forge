package main

import "github.com/byte-v-forge/common-lib/eventcatalog"

var (
	mailboxInboxFetchRequested   = eventcatalog.MailboxInboxFetchRequested
	mailboxRegistrationRequested = eventcatalog.MailboxRegistrationRequested
	mailboxOAuthRequested        = eventcatalog.MailboxOAuthRequested
)
