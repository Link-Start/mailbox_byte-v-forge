package mailboxpg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (q *inboxMessageQuery) OrderByLatest() *inboxMessageQuery {
	q.orderBy = "received_at DESC, updated_at DESC, message_key DESC"
	return q
}

func (q *inboxMessageQuery) Limit(limit int) *inboxMessageQuery {
	q.limit = limit
	return q
}

func (q *inboxMessageQuery) Offset(offset int) *inboxMessageQuery {
	q.offset = offset
	return q
}

func (q *inboxMessageQuery) Query(ctx context.Context, querier mailboxQuerier) (pgx.Rows, error) {
	query, args := q.SQL()
	return querier.Query(ctx, query, args...)
}

func (q *inboxMessageQuery) SQL() (string, []any) {
	query := inboxMessageSelectSQL
	if len(q.conditions) > 0 {
		query += "WHERE " + strings.Join(q.conditions, " AND ")
	}
	if strings.TrimSpace(q.orderBy) != "" {
		query += " ORDER BY " + q.orderBy
	}
	args := append([]any{}, q.args...)
	if q.limit > 0 {
		args = append(args, q.limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	if q.offset > 0 {
		args = append(args, q.offset)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}
	return query, args
}

func (q *inboxMessageQuery) addArg(value any) string {
	q.args = append(q.args, value)
	return fmt.Sprintf("$%d", len(q.args))
}
