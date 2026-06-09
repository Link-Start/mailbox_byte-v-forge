package mailboxpg

import (
	"context"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

func (r *Repository) listStoredInboxOnlyVirtualMailboxes(ctx context.Context, provider string, filter mailboxprovider.ListQuery, normalizeProvider func(string) string, prepareProjection func(*mailboxmodel.Record)) ([]*mailboxmodel.Record, error) {
	provider = mailboxprovider.NormalizeKey(provider)
	if r == nil || r.pool == nil || provider == "" {
		return []*mailboxmodel.Record{}, nil
	}
	if normalizeProvider == nil {
		normalizeProvider = mailboxprovider.NormalizeKey
	}
	if prepareProjection == nil {
		prepareProjection = func(*mailboxmodel.Record) {}
	}
	query := newVirtualMailboxQuery(provider)
	if filter.EmailAddress != "" {
		query.WhereEmail(filter.EmailAddress)
	}
	if filter.HasCursor() {
		query.WhereCursor(filter.Cursor)
	}
	rows, err := query.Limit(filter.ScanLimit()).Query(ctx, r.pool)
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
		out = append(out, row.ToRecord(normalizeProvider, prepareProjection))
	}
	return out, rows.Err()
}
