package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/byte-v-forge/common-lib/envx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxpg"
)

type oauthEntry struct {
	refreshToken string
	manager      *OAuthManager
}

type MailWatcher struct {
	inbox        *inboxapp.Service
	mailboxes    *mailboxpg.Repository
	messageLimit int
	pollInterval int
	inboxOverlap int
	httpClient   *http.Client
	events       *mailboxHotStream

	mu            sync.Mutex
	oauthManagers map[string]oauthEntry
}

type GraphFetchError struct {
	StatusCode int
	Body       string
	RetryAfter time.Duration
}

func (e *GraphFetchError) Error() string {
	body := safeMailboxText(e.Body)
	if len(body) > 500 {
		body = body[:500]
	}
	return fmt.Sprintf("status=%d body=%s", e.StatusCode, body)
}

func (e *GraphFetchError) IsAuth() bool {
	return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
}

func (e *GraphFetchError) Retryable() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= http.StatusInternalServerError
}

func NewMailWatcher(inbox *inboxapp.Service, mailboxes *mailboxpg.Repository, events *mailboxHotStream) *MailWatcher {
	messageLimit := envx.Int("OUTLOOK_MESSAGE_LIMIT", defaultMessageLimit)
	if messageLimit < 1 {
		messageLimit = 1
	}
	if messageLimit > 100 {
		messageLimit = 100
	}
	pollInterval := envx.Int("OUTLOOK_POLL_INTERVAL_SECONDS", defaultPollIntervalSeconds)
	if pollInterval < 1 {
		pollInterval = 1
	}
	inboxOverlap := envx.Int("OUTLOOK_INBOX_OVERLAP_SECONDS", defaultInboxOverlapSeconds)
	if inboxOverlap < 0 {
		inboxOverlap = 0
	}
	timeout := envx.Int("OUTLOOK_HTTP_TIMEOUT_SECONDS", defaultHTTPTimeoutSeconds)
	if timeout <= 0 {
		timeout = defaultHTTPTimeoutSeconds
	}
	return &MailWatcher{
		inbox:         inbox,
		mailboxes:     mailboxes,
		messageLimit:  messageLimit,
		pollInterval:  pollInterval,
		inboxOverlap:  inboxOverlap,
		httpClient:    &http.Client{Timeout: time.Duration(timeout) * time.Second},
		events:        events,
		oauthManagers: map[string]oauthEntry{},
	}
}

func (w *MailWatcher) PollForEmail(ctx context.Context, email string) error {
	mailbox, err := w.mailboxes.PollMailboxForEmail(ctx, email)
	if err != nil {
		return err
	}
	messages, err := w.fetchMailboxMessages(ctx, mailbox, w.messageLimit, 0)
	if err != nil {
		return err
	}
	unseen, err := w.inbox.RecordMessages(ctx, emailProviderOutlook, inboxMessages(mailbox.GetEmailAddress(), messages), true)
	if err != nil {
		return err
	}
	w.DispatchMailboxEvents(ctx, unseen)
	return nil
}

func (w *MailWatcher) FetchMailboxInbox(ctx context.Context, mailbox *mailboxmodel.Record, limit int32, receivedAfterUnix int64) ([]*mailboxv1.EmailInboxMessage, error) {
	watermark, err := w.inbox.InboxWatermark(ctx, mailbox.GetEmailAddress())
	if err != nil {
		return nil, err
	}
	messageLimit := inboxapp.MessageLimitValue(limit, w.messageLimit)
	receivedAfter := inboxapp.InboxReceivedAfter(watermark, w.inboxOverlap)
	hasPersistedMessages, err := w.inbox.HasMessages(ctx, mailbox.GetEmailAddress())
	if err != nil {
		return nil, err
	}
	if !hasPersistedMessages {
		receivedAfter = 0
	}
	messages, err := w.fetchMailboxMessages(ctx, mailbox, messageLimit, receivedAfter)
	if err != nil {
		return nil, err
	}
	unseen, err := w.inbox.RecordMessages(ctx, emailProviderOutlook, inboxMessages(mailbox.GetEmailAddress(), messages), true)
	if err != nil {
		return nil, err
	}
	w.DispatchMailboxEvents(ctx, unseen)
	return w.inbox.ListMessagesSince(ctx, mailbox.GetEmailAddress(), int32(messageLimit), receivedAfterUnix)
}

func (w *MailWatcher) DispatchMailboxEvents(ctx context.Context, messages []*mailboxv1.EmailInboxMessage) {
	if len(messages) == 0 {
		return
	}
	if w.events != nil {
		w.events.PublishEmailMessages(ctx, messages)
	}
}
