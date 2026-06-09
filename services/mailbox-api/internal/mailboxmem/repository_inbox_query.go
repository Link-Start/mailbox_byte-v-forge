package mailboxmem

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
)

func (r *Repository) InboxWatermark(ctx context.Context, email string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	email = emailx.Normalize(email)
	if email == "" {
		return 0, errors.New("email_address is required")
	}
	r.mu.RLock()
	entry, ok := r.mailboxes[email]
	r.mu.RUnlock()
	if !ok || entry.record == nil {
		return 0, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	return entry.inboxWatermark, nil
}

func (r *Repository) HasInboxMessages(ctx context.Context, email string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	email = emailx.Normalize(email)
	if email == "" {
		return false, errors.New("email_address is required")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, message := range r.messages {
		if message.row.MailboxEmail == email {
			return true, nil
		}
	}
	return false, nil
}

func (r *Repository) ListInboxRows(ctx context.Context, email string, limit int, receivedAfterUnix int64) ([]inboxapp.MessageRow, error) {
	return r.filterInboxRows(ctx, email, "", receivedAfterUnix, false, limit)
}

func (r *Repository) GetInboxRow(ctx context.Context, email string, messageID string, provider string) (inboxapp.MessageRow, bool, error) {
	if err := ctx.Err(); err != nil {
		return inboxapp.MessageRow{}, false, err
	}
	email = emailx.Normalize(email)
	if email == "" {
		return inboxapp.MessageRow{}, false, errors.New("email_address is required")
	}
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return inboxapp.MessageRow{}, false, errors.New("message_id is required")
	}
	provider = r.providers.NormalizeProviderInput(provider)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, message := range r.messages {
		if message.row.MailboxEmail != email || message.row.ID != messageID {
			continue
		}
		if provider != "" && message.row.Provider != provider {
			continue
		}
		return message.row, true, nil
	}
	return inboxapp.MessageRow{}, false, nil
}

func (r *Repository) LatestInboxRows(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, limit int) ([]inboxapp.MessageRow, error) {
	return r.filterInboxRows(ctx, email, subjectKeyword, issuedAfterUnix, true, limit)
}
