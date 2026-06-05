package main

import (
	"context"

	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxpg"
)

type MailWatcher struct {
	inbox     *inboxapp.Service
	mailboxes *mailboxpg.Repository
	sources   *mailboxInboxSourceRegistry
	events    *mailboxHotStream
}

func NewMailWatcher(inbox *inboxapp.Service, mailboxes *mailboxpg.Repository, sources *mailboxInboxSourceRegistry, events *mailboxHotStream) *MailWatcher {
	return &MailWatcher{
		inbox:     inbox,
		mailboxes: mailboxes,
		sources:   sources,
		events:    events,
	}
}

func (w *MailWatcher) PollForEmail(ctx context.Context, email string) error {
	mailbox, err := w.mailboxes.PollMailboxForEmail(ctx, email)
	if err != nil {
		return err
	}
	source, err := w.sources.SourceForMailbox(mailbox)
	if err != nil {
		return err
	}
	messages, err := source.FetchInboxMessages(ctx, mailbox, source.DefaultMessageLimit(), 0)
	if err != nil {
		return err
	}
	unseen, err := w.inbox.RecordMessages(ctx, source.ProviderKey(), messages, true)
	if err != nil {
		return err
	}
	w.DispatchMailboxEvents(ctx, unseen)
	return nil
}

func (w *MailWatcher) FetchMailboxInbox(ctx context.Context, mailbox *mailboxmodel.Record, limit int32, receivedAfterUnix int64) ([]*mailboxv1.EmailInboxMessage, error) {
	source, err := w.sources.SourceForMailbox(mailbox)
	if err != nil {
		return nil, err
	}
	watermark, err := w.inbox.InboxWatermark(ctx, mailbox.GetEmailAddress())
	if err != nil {
		return nil, err
	}
	messageLimit := inboxapp.MessageLimitValue(limit, source.DefaultMessageLimit())
	receivedAfter := inboxapp.InboxReceivedAfter(watermark, source.InboxOverlap())
	hasPersistedMessages, err := w.inbox.HasMessages(ctx, mailbox.GetEmailAddress())
	if err != nil {
		return nil, err
	}
	if !hasPersistedMessages {
		receivedAfter = 0
	}
	messages, err := source.FetchInboxMessages(ctx, mailbox, messageLimit, receivedAfter)
	if err != nil {
		return nil, err
	}
	unseen, err := w.inbox.RecordMessages(ctx, source.ProviderKey(), messages, true)
	if err != nil {
		return nil, err
	}
	w.DispatchMailboxEvents(ctx, unseen)
	return w.inbox.ListMessagesSince(ctx, mailbox.GetEmailAddress(), int32(messageLimit), receivedAfterUnix)
}

func (w *MailWatcher) DefaultMessageLimit() int {
	if w == nil || w.sources == nil {
		return defaultMessageLimit
	}
	return w.sources.DefaultMessageLimit()
}

func (w *MailWatcher) DefaultPollInterval() int {
	if w == nil || w.sources == nil {
		return defaultPollIntervalSeconds
	}
	return w.sources.DefaultPollInterval()
}

func (w *MailWatcher) DispatchMailboxEvents(ctx context.Context, messages []*mailboxv1.EmailInboxMessage) {
	if len(messages) == 0 {
		return
	}
	if w.events != nil {
		w.events.PublishEmailMessages(ctx, messages)
	}
}
