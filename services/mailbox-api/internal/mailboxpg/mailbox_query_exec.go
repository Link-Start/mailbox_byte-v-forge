package mailboxpg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (q *mailboxSelectQuery) Query(ctx context.Context, querier mailboxQuerier) (pgx.Rows, error) {
	query, args, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return querier.Query(ctx, query, args...)
}

func (q *mailboxSelectQuery) ScanOne(ctx context.Context, querier mailboxQuerier) (*MailboxRow, error) {
	query, args, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return ScanMailbox(querier.QueryRow(ctx, query, args...))
}

func (q *mailboxSelectQuery) SQL() (string, []any, error) {
	if q.err != nil {
		return "", nil, q.err
	}
	query, err := q.selectSQL()
	if err != nil {
		return "", nil, err
	}
	if len(q.conditions) > 0 {
		query += " WHERE " + strings.Join(q.conditions, " AND ")
	}
	if strings.TrimSpace(q.orderBy) != "" {
		query += " ORDER BY " + q.orderBy
	}
	args := append([]any{}, q.args...)
	if q.limit > 0 {
		args = append(args, q.limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	return query, args, nil
}

func (q *mailboxSelectQuery) addArg(value any) string {
	q.args = append(q.args, value)
	return fmt.Sprintf("$%d", len(q.args))
}
