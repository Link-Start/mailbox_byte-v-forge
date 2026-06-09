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
	row, err := r.pollMailboxRow(ctx, email)
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

func (r *Repository) pollMailboxRow(ctx context.Context, email string) (*MailboxRow, error) {
	row, err := r.newMailboxSelectQuery().WhereEmail(email).ScanOne(ctx, r.pool)
	if !errors.Is(err, pgx.ErrNoRows) {
		return row, err
	}
	canonical := emailx.CanonicalPlusAlias(email)
	if canonical == "" || canonical == email {
		return row, err
	}
	return r.newMailboxSelectQuery().WhereEmail(canonical).ScanOne(ctx, r.pool)
}
