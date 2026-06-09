package mailboxmem

import (
	"context"
	"fmt"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
)

func (r *Repository) FindMailbox(ctx context.Context, email string) (*mailboxmodel.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	email = emailx.Normalize(email)
	r.mu.RLock()
	record := cloneRecord(r.mailboxes[email].record)
	r.mu.RUnlock()
	if record == nil {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	return r.project(record), nil
}

func (r *Repository) PollMailboxForEmail(ctx context.Context, email string) (*mailboxmodel.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	email = emailx.Normalize(email)
	record := r.findPollRecord(email)
	if record == nil {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	if err := r.providers.ValidatePoll(providerRecord(record)); err != nil {
		return nil, err
	}
	return r.project(record), nil
}

func (r *Repository) findPollRecord(email string) *mailboxmodel.Record {
	r.mu.RLock()
	defer r.mu.RUnlock()
	record := cloneRecord(r.mailboxes[email].record)
	if record != nil {
		return record
	}
	canonical := emailx.CanonicalPlusAlias(email)
	if canonical == "" || canonical == email {
		return nil
	}
	return cloneRecord(r.mailboxes[canonical].record)
}
