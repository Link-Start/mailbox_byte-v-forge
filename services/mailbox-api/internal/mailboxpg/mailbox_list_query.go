package mailboxpg

import (
	"context"
	"sort"
	"strings"
	"time"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/pagex"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

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
