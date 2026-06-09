package mailboxmem

import (
	"context"
	"errors"
	"time"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/stringx"
)

func (r *Repository) UpsertMailbox(ctx context.Context, mailbox *mailboxmodel.Record) (*mailboxmodel.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if mailbox == nil {
		return nil, errors.New("mailbox is required")
	}
	email := emailx.Normalize(mailbox.GetEmailAddress())
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	requestedProvider := r.providers.NormalizeProviderInput(mailbox.GetProviderKey())
	now := time.Now().Unix()

	r.mu.Lock()
	entry, exists := r.mailboxes[email]
	record := cloneRecord(entry.record)
	if record == nil {
		record = &mailboxmodel.Record{EmailAddress: email, CreatedAt: now}
	}
	if requestedProvider != "" || record.ProviderKey == "" {
		record.ProviderKey = stringx.FirstNonEmpty(requestedProvider, r.providers.DefaultKey())
	}
	mergeCredentials(record, mailbox, exists)
	record.EmailAddress = email
	record.Domain = domainForEmail(email)
	record.UpdatedAt = now
	entry.record = record
	r.mailboxes[email] = entry
	r.mu.Unlock()

	return r.FindMailbox(ctx, email)
}
