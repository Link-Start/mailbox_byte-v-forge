package mailboxpg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"

	"mailboxapi/internal/mailboxmodel"
)

func (r *Repository) FindMailbox(ctx context.Context, email string) (*mailboxmodel.Record, error) {
	row, err := r.newMailboxSelectQuery().WhereEmail(email).ScanOne(ctx, r.pool)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	if err != nil {
		return nil, err
	}
	return r.recordFromRow(row), nil
}

func (r *Repository) PollMailboxForEmail(ctx context.Context, email string) (*mailboxmodel.Record, error) {
	email = emailx.Normalize(email)
	row, err := r.newMailboxSelectQuery().WhereEmail(email).ScanOne(ctx, r.pool)
	if errors.Is(err, pgx.ErrNoRows) {
		canonical := emailx.CanonicalPlusAlias(email)
		if canonical != "" && canonical != email {
			row, err = r.newMailboxSelectQuery().WhereEmail(canonical).ScanOne(ctx, r.pool)
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	if err != nil {
		return nil, err
	}
	if err := r.providers.ValidatePoll(row.ToProviderRecord()); err != nil {
		return nil, err
	}
	return r.recordFromRow(row), nil
}

func (r *Repository) ListMailboxes(ctx context.Context, authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxmodel.ListPage, error) {
	query, err := r.newMailboxListQuery(authStatus, provider, emailAddress, cursorValue, limit)
	if err != nil {
		return mailboxmodel.ListPage{}, err
	}
	stored, err := r.listStoredMailboxes(ctx, query)
	if err != nil {
		return mailboxmodel.ListPage{}, err
	}
	rows := stored
	virtual, err := r.listVirtualMailboxes(ctx, query)
	if err != nil {
		return mailboxmodel.ListPage{}, err
	}
	if len(virtual) > 0 {
		rows = append(rows, virtual...)
	}
	return mailboxPageFromRows(rows, query.Limit), nil
}

func (r *Repository) ListOAuthMailboxes(ctx context.Context, limit int32) ([]*mailboxmodel.Record, error) {
	n := int(limit)
	if n <= 0 {
		n = 100
	}
	if n > 500 {
		n = 500
	}
	rows, err := r.newMailboxSelectQuery().
		WhereAuthStatus("", mailboxmodel.AuthStatusAuthorized).
		OrderByUpdatedDesc().
		Limit(n).
		Query(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*mailboxmodel.Record{}
	for rows.Next() {
		row, err := ScanMailbox(rows)
		if err != nil {
			return nil, err
		}
		if r.providers.ValidatePoll(row.ToProviderRecord()) != nil {
			continue
		}
		out = append(out, r.recordFromRow(row))
	}
	return out, rows.Err()
}
