package mailboxpg

import (
	"context"

	"mailboxapi/internal/mailboxmodel"
)

func (r *Repository) ListOAuthMailboxes(ctx context.Context, limit int32) ([]*mailboxmodel.Record, error) {
	rows, err := r.newMailboxSelectQuery().
		WhereAuthStatus("", mailboxmodel.AuthStatusAuthorized).
		OrderByUpdatedDesc().
		Limit(oauthMailboxLimit(limit)).
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

func oauthMailboxLimit(limit int32) int {
	n := int(limit)
	if n <= 0 {
		n = 100
	}
	if n > 500 {
		n = 500
	}
	return n
}
