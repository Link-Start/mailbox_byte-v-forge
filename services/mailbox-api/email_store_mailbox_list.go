package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/accountmodel"
	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/byte-v-forge/common-lib/pagex"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxpg"
	"mailboxapi/internal/mailboxprovider"
)

func (s *MailboxStore) ListMailboxes(ctx context.Context, authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxmodel.ListPage, error) {
	query, err := s.newMailboxListQuery(authStatus, provider, emailAddress, cursorValue, limit)
	if err != nil {
		return mailboxmodel.ListPage{}, err
	}
	stored, err := s.listStoredMailboxes(ctx, query)
	if err != nil {
		return mailboxmodel.ListPage{}, err
	}
	rows := stored
	virtual, err := s.providers.VirtualMailboxes(ctx, s.pool, query)
	if err != nil {
		return mailboxmodel.ListPage{}, err
	}
	if len(virtual) > 0 {
		rows = append(rows, virtual...)
	}
	return mailboxPageFromRows(rows, query.Limit), nil
}

func (s *MailboxStore) newMailboxListQuery(authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxprovider.ListQuery, error) {
	cursor, err := pagex.DecodeKeysetCursor(cursorValue)
	if err != nil {
		return mailboxprovider.ListQuery{}, mailboxmodel.ErrInvalidMailboxListCursor
	}
	return mailboxprovider.ListQuery{
		AuthStatus:   strings.TrimSpace(authStatus),
		Provider:     s.providers.NormalizeProviderInput(provider),
		EmailAddress: emailx.Normalize(emailAddress),
		Cursor:       cursor,
		Limit:        accountmodel.NormalizePageLimit(int(limit)),
	}, nil
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

func (s *MailboxStore) listStoredMailboxes(ctx context.Context, filter mailboxprovider.ListQuery) ([]*mailboxmodel.Record, error) {
	args := []any{}
	query := s.providers.MailboxSelectSQL() + ` WHERE 1=1`
	if filter.AuthStatus != "" {
		query += " AND " + s.providers.AuthFilter(filter.Provider, filter.AuthStatus, &args)
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
		row, err := mailboxpg.ScanMailbox(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s.recordFromMailboxRow(row))
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
	query := s.providers.MailboxSelectSQL() + " WHERE " + s.providers.AuthFilter("", mailboxmodel.AuthStatusAuthorized, &args)
	args = append(args, n)
	query += fmt.Sprintf(" ORDER BY m.updated_at DESC, m.email DESC LIMIT $%d", len(args))
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*mailboxmodel.Record{}
	for rows.Next() {
		row, err := mailboxpg.ScanMailbox(rows)
		if err != nil {
			return nil, err
		}
		if s.providers.ValidatePoll(row.ToProviderRecord()) != nil {
			continue
		}
		out = append(out, s.recordFromMailboxRow(row))
	}
	return out, rows.Err()
}
