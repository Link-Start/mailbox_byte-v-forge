package mailboxpg

import (
	"context"
	"fmt"

	"github.com/byte-v-forge/common-lib/emailx"

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
	args := []any{provider}
	where := ""
	if filter.EmailAddress != "" {
		args = append(args, filter.EmailAddress)
		where += fmt.Sprintf(" AND msg.mailbox_email = $%d", len(args))
	}
	query := fmt.Sprintf(`
		SELECT $1 || ':' || v.mailbox_email, v.mailbox_email,
			$1, '', '', '', '', '', v.created_at, v.updated_at
		FROM (
			SELECT msg.mailbox_email, MIN(msg.created_at) AS created_at, MAX(msg.updated_at) AS updated_at
			FROM mailbox_inbox_messages msg
			WHERE msg.provider = $1
			  AND NOT EXISTS (SELECT 1 FROM mailboxes m WHERE m.email = msg.mailbox_email)
			  %s
			GROUP BY msg.mailbox_email
		) v
		WHERE 1=1
	`, where)
	if filter.HasCursor() {
		args = append(args, filter.Cursor.UpdatedAt.Unix(), emailx.Normalize(filter.Cursor.ID))
		query += fmt.Sprintf(" AND (v.updated_at < $%d OR (v.updated_at = $%d AND v.mailbox_email < $%d))", len(args)-1, len(args)-1, len(args))
	}
	args = append(args, filter.ScanLimit())
	query += fmt.Sprintf(" ORDER BY v.updated_at DESC, v.mailbox_email DESC LIMIT $%d", len(args))

	rows, err := r.pool.Query(ctx, query, args...)
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

func (r *Repository) listVirtualMailboxes(ctx context.Context, query mailboxprovider.ListQuery) ([]*mailboxmodel.Record, error) {
	out := []*mailboxmodel.Record{}
	for _, source := range r.providers.VirtualMailboxSources() {
		if query.Provider != "" && query.Provider != source.Key() {
			continue
		}
		if !source.StoredInboxOnly() || !source.IncludeVirtual(query.AuthStatus) {
			continue
		}
		items, err := r.listStoredInboxOnlyVirtualMailboxes(ctx, source.Key(), query, r.providers.NormalizeProviderInput, r.providers.PrepareProjection)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}
