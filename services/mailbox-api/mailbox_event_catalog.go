package main

import "mailboxapi/internal/eventcatalog"

var (
	mailboxInboxFetchRequested = eventcatalog.Definition{
		Subject:          "mailbox.inbox.fetch.requested",
		EventName:        "mailbox.inbox.fetch_requested",
		EventVersion:     eventcatalog.EventVersionV1,
		Kind:             eventcatalog.KindCommand,
		PayloadType:      "byte.v.forge.mailbox.internal.v1.MailboxInboxFetchRequest",
		OwnerService:     "mailbox-api",
		ConsumerDurable:  "mailbox-inbox-fetch",
		Retryable:        true,
		MaxDeliveries:    20,
		RetryDelaySecond: 5,
	}
	mailboxRegistrationRequested = eventcatalog.Definition{
		Subject:          "mailbox.registration.requested",
		EventName:        "mailbox.registration.requested",
		EventVersion:     eventcatalog.EventVersionV1,
		Kind:             eventcatalog.KindCommand,
		PayloadType:      "byte.v.forge.mailbox.internal.v1.MailboxRegistrationOperationRequest",
		OwnerService:     "mailbox-api",
		ConsumerDurable:  "mailbox-registration",
		Retryable:        true,
		MaxDeliveries:    20,
		RetryDelaySecond: 5,
	}
	mailboxOAuthRequested = eventcatalog.Definition{
		Subject:          "mailbox.oauth.requested",
		EventName:        "mailbox.oauth.requested",
		EventVersion:     eventcatalog.EventVersionV1,
		Kind:             eventcatalog.KindCommand,
		PayloadType:      "byte.v.forge.mailbox.internal.v1.MailboxOAuthOperationRequest",
		OwnerService:     "mailbox-api",
		ConsumerDurable:  "mailbox-oauth",
		Retryable:        true,
		MaxDeliveries:    20,
		RetryDelaySecond: 5,
	}
)
