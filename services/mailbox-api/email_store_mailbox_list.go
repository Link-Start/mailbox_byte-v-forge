package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/accountmodel"
	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/byte-v-forge/common-lib/pagex"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

var errInvalidMailboxListCursor = errors.New("invalid mailbox cursor")

type mailboxListPage struct {
	Mailboxes  []*mailboxmodel.Record
	NextCursor string
}

func (s *MailboxStore) ListMailboxes(ctx context.Context, authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxListPage, error) {
	query, err := newMailboxListQuery(authStatus, provider, emailAddress, cursorValue, limit)
	if err != nil {
		return mailboxListPage{}, err
	}
	stored, err := s.listStoredMailboxes(ctx, query)
	if err != nil {
		return mailboxListPage{}, err
	}
	rows := stored
	virtual, err := listMailboxProviderVirtualMailboxes(ctx, s.pool, query)
	if err != nil {
		return mailboxListPage{}, err
	}
	if len(virtual) > 0 {
		rows = append(rows, virtual...)
	}
	return mailboxPageFromRows(rows, query.Limit), nil
}

func newMailboxListQuery(authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxprovider.ListQuery, error) {
	cursor, err := pagex.DecodeKeysetCursor(cursorValue)
	if err != nil {
		return mailboxprovider.ListQuery{}, errInvalidMailboxListCursor
	}
	return mailboxprovider.ListQuery{
		AuthStatus:   strings.TrimSpace(authStatus),
		Provider:     normalizeEmailProvider(provider),
		EmailAddress: emailx.Normalize(emailAddress),
		Cursor:       cursor,
		Limit:        accountmodel.NormalizePageLimit(int(limit)),
	}, nil
}

func mailboxPageFromRows(rows []*mailboxmodel.Record, limit int) mailboxListPage {
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
	return mailboxListPage{Mailboxes: page.Items, NextCursor: page.NextCursor}
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

func (s *MailboxStore) listStoredMailboxes(ctx context.Context, filter mailboxprovider.ListQuery) ([]*mailboxmodel.Record, error) {
	args := []any{}
	query := mailboxSelectSQL() + ` WHERE 1=1`
	if filter.AuthStatus != "" {
		query += " AND " + mailboxProviderAuthFilter(filter.Provider, filter.AuthStatus, &args)
	}
	if filter.Provider != "" {
		args = append(args, filter.Provider)
		query += fmt.Sprintf(" AND m.provider = $%d", len(args))
	}
	if filter.EmailAddress != "" {
		args = append(args, filter.EmailAddress)
		query += fmt.Sprintf(" AND m.email = $%d", len(args))
	}
	if filter.HasCursor() {
		args = append(args, filter.Cursor.UpdatedAt.Unix(), emailx.Normalize(filter.Cursor.ID))
		query += fmt.Sprintf(" AND (m.updated_at < $%d OR (m.updated_at = $%d AND m.email < $%d))", len(args)-1, len(args)-1, len(args))
	}
	args = append(args, filter.ScanLimit())
	query += fmt.Sprintf(" ORDER BY m.updated_at DESC, m.email DESC LIMIT $%d", len(args))

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*mailboxmodel.Record{}
	for rows.Next() {
		row, err := scanMailbox(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, row.toRecord())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *MailboxStore) ListOAuthMailboxes(ctx context.Context, limit int32) ([]*mailboxmodel.Record, error) {
	n := int(limit)
	if n <= 0 {
		n = 100
	}
	if n > 500 {
		n = 500
	}
	args := []any{}
	query := mailboxSelectSQL() + " WHERE " + mailboxProviderAuthFilter("", authStatusAuthorized, &args)
	args = append(args, n)
	query += fmt.Sprintf(" ORDER BY m.updated_at DESC, m.email DESC LIMIT $%d", len(args))
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*mailboxmodel.Record{}
	for rows.Next() {
		row, err := scanMailbox(rows)
		if err != nil {
			return nil, err
		}
		if mailboxProviderValidatePoll(row) != nil {
			continue
		}
		out = append(out, row.toRecord())
	}
	return out, rows.Err()
}
