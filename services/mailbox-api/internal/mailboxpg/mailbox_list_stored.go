package mailboxpg

import (
	"context"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

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
