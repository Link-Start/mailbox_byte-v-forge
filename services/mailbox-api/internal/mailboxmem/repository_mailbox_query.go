package mailboxmem

import (
	"context"
	"fmt"
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
	"mailboxapi/internal/pagex"
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
	r.mu.RLock()
	record := cloneRecord(r.mailboxes[email].record)
	if record == nil {
		canonical := emailx.CanonicalPlusAlias(email)
		if canonical != "" && canonical != email {
			record = cloneRecord(r.mailboxes[canonical].record)
		}
	}
	r.mu.RUnlock()
	if record == nil {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	if err := r.providers.ValidatePoll(providerRecord(record)); err != nil {
		return nil, err
	}
	return r.project(record), nil
}

func (r *Repository) ListMailboxes(ctx context.Context, authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxmodel.ListPage, error) {
	if err := ctx.Err(); err != nil {
		return mailboxmodel.ListPage{}, err
	}
	cursor, err := pagex.DecodeKeysetCursor(cursorValue)
	if err != nil {
		return mailboxmodel.ListPage{}, mailboxmodel.ErrInvalidMailboxListCursor
	}
	query := mailboxprovider.ListQuery{
		AuthStatus:   strings.TrimSpace(authStatus),
		Provider:     r.providers.NormalizeProviderInput(provider),
		EmailAddress: emailx.Normalize(emailAddress),
		Cursor:       cursor,
		Limit:        pagex.NormalizePageLimit(int(limit)),
	}
	r.mu.RLock()
	rows := r.listStoredMailboxesLocked(query)
	rows = append(rows, r.listVirtualMailboxesLocked(query)...)
	r.mu.RUnlock()
	return mailboxPageFromRows(rows, query.Limit), nil
}

func (r *Repository) ListOAuthMailboxes(ctx context.Context, limit int32) ([]*mailboxmodel.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n := int(limit)
	if n <= 0 {
		n = 100
	}
	if n > 500 {
		n = 500
	}
	r.mu.RLock()
	rows := make([]*mailboxmodel.Record, 0, len(r.mailboxes))
	for _, entry := range r.mailboxes {
		record := cloneRecord(entry.record)
		if record == nil || record.AuthStatus != mailboxmodel.AuthStatusAuthorized {
			continue
		}
		if r.providers.ValidatePoll(providerRecord(record)) != nil {
			continue
		}
		rows = append(rows, r.project(record))
	}
	r.mu.RUnlock()
	sortMailboxRows(rows)
	if len(rows) > n {
		rows = rows[:n]
	}
	return rows, nil
}
