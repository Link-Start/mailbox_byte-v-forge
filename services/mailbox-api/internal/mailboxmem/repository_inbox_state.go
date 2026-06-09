package mailboxmem

import (
	"context"
	"errors"
	"fmt"

	"mailboxapi/internal/emailx"
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
