package mailboxpg

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/pagex"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
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

func (r *Repository) newMailboxListQuery(authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxprovider.ListQuery, error) {
	cursor, err := pagex.DecodeKeysetCursor(cursorValue)
	if err != nil {
		return mailboxprovider.ListQuery{}, mailboxmodel.ErrInvalidMailboxListCursor
	}
	return mailboxprovider.ListQuery{
		AuthStatus:   strings.TrimSpace(authStatus),
		Provider:     r.providers.NormalizeProviderInput(provider),
		EmailAddress: emailx.Normalize(emailAddress),
		Cursor:       cursor,
		Limit:        pagex.NormalizePageLimit(int(limit)),
	}, nil
}

func (r *Repository) listStoredMailboxes(ctx context.Context, filter mailboxprovider.ListQuery) ([]*mailboxmodel.Record, error) {
	query := r.newMailboxSelectQuery()
	if filter.AuthStatus != "" {
		query.WhereAuthStatus(filter.Provider, filter.AuthStatus)
	}
	if filter.Provider != "" {
		query.WhereProvider(filter.Provider)
	}
	if filter.EmailAddress != "" {
		query.WhereEmail(filter.EmailAddress)
	}
	if filter.HasCursor() {
		query.WhereCursor(filter.Cursor)
	}
	rows, err := query.OrderByUpdatedDesc().Limit(filter.ScanLimit()).Query(ctx, r.pool)
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
		out = append(out, r.recordFromRow(row))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func mailboxPageFromRows(rows []*mailboxmodel.Record, limit int) mailboxmodel.ListPage {
	rows = uniqueMailboxRows(rows)
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].GetUpdatedAt() == rows[j].GetUpdatedAt() {
			return rows[i].GetEmailAddress() > rows[j].GetEmailAddress()
		}
		return rows[i].GetUpdatedAt() > rows[j].GetUpdatedAt()
	})
	page := pagex.NewKeysetPage(rows, limit, func(mailbox *mailboxmodel.Record) pagex.KeysetCursor {
		return pagex.KeysetCursorValue(time.Unix(mailbox.GetUpdatedAt(), 0).UTC(), mailbox.GetEmailAddress())
	})
	return mailboxmodel.ListPage{Mailboxes: page.Items, NextCursor: page.NextCursor}
}

func uniqueMailboxRows(rows []*mailboxmodel.Record) []*mailboxmodel.Record {
	seen := map[string]struct{}{}
	out := make([]*mailboxmodel.Record, 0, len(rows))
	for _, row := range rows {
		email := emailx.Normalize(row.GetEmailAddress())
		if email == "" {
			continue
		}
		if _, exists := seen[email]; exists {
			continue
		}
		seen[email] = struct{}{}
		out = append(out, row)
	}
	return out
}
