package main

import "github.com/byte-v-forge/common-lib/eventcatalog"

var (
	mailboxInboxFetchRequested = eventcatalog.Definition{
		Subject:          "byte.v.forge.mailbox.inbox.fetch.requested",
		EventName:        "mailbox.inbox.fetch_requested",
		EventVersion:     mailboxPlatformEventVersion,
		Kind:             eventcatalog.KindCommand,
		PayloadType:      "mailbox.MailboxInboxFetchRequest",
		OwnerService:     mailboxPlatformEventSource,
		ConsumerDurable:  "mailbox-inbox-fetch",
		Retryable:        true,
		MaxDeliveries:    20,
		RetryDelaySecond: 5,
	}
	mailboxRegistrationRequested = eventcatalog.Definition{
		Subject:          "byte.v.forge.mailbox.registration.requested",
		EventName:        "mailbox.registration.requested",
		EventVersion:     mailboxPlatformEventVersion,
		Kind:             eventcatalog.KindCommand,
		PayloadType:      "mailbox.MailboxRegistrationOperationRequest",
		OwnerService:     mailboxPlatformEventSource,
		ConsumerDurable:  "mailbox-registration",
		Retryable:        true,
		MaxDeliveries:    20,
		RetryDelaySecond: 5,
	}
	mailboxOAuthRequested = eventcatalog.Definition{
		Subject:          "byte.v.forge.mailbox.oauth.requested",
		EventName:        "mailbox.oauth.requested",
		EventVersion:     mailboxPlatformEventVersion,
		Kind:             eventcatalog.KindCommand,
		PayloadType:      "mailbox.MailboxOAuthOperationRequest",
		OwnerService:     mailboxPlatformEventSource,
		ConsumerDurable:  "mailbox-oauth",
		Retryable:        true,
		MaxDeliveries:    20,
		RetryDelaySecond: 5,
	}
)
