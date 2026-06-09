package mailboxpg

import (
	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/inboxapp"
)

func scanInboxRows(rows pgx.Rows) ([]inboxapp.MessageRow, error) {
	defer rows.Close()

	out := []inboxapp.MessageRow{}
	for rows.Next() {
		row, err := scanInboxMessageRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func normalizeInboxRowLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	return limit
}
