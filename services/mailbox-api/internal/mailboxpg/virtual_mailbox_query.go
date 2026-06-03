package mailboxpg

import (
	"context"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/byte-v-forge/common-lib/pagex"
	"github.com/jackc/pgx/v5"
)

type virtualMailboxQuery struct {
	innerConditions []string
	outerConditions []string
	args            []any
	limit           int
}

func newVirtualMailboxQuery(provider string) *virtualMailboxQuery {
	return &virtualMailboxQuery{args: []any{provider}}
}

func (q *virtualMailboxQuery) WhereEmail(email string) *virtualMailboxQuery {
	q.innerConditions = append(q.innerConditions, fmt.Sprintf("msg.mailbox_email = %s", q.addArg(emailx.Normalize(email))))
	return q
}

func (q *virtualMailboxQuery) WhereCursor(cursor pagex.KeysetCursor) *virtualMailboxQuery {
	updatedAtParam := q.addArg(cursor.UpdatedAt.Unix())
	emailParam := q.addArg(emailx.Normalize(cursor.ID))
	q.outerConditions = append(q.outerConditions, fmt.Sprintf("(v.updated_at < %s OR (v.updated_at = %s AND v.mailbox_email < %s))", updatedAtParam, updatedAtParam, emailParam))
	return q
}

func (q *virtualMailboxQuery) Limit(limit int) *virtualMailboxQuery {
	q.limit = limit
	return q
}

func (q *virtualMailboxQuery) Query(ctx context.Context, querier mailboxQuerier) (pgx.Rows, error) {
	query, args := q.SQL()
	return querier.Query(ctx, query, args...)
}

func (q *virtualMailboxQuery) SQL() (string, []any) {
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
	`, q.innerWhereSQL())
	if len(q.outerConditions) > 0 {
		query += " WHERE " + strings.Join(q.outerConditions, " AND ")
	}
	args := append([]any{}, q.args...)
	if q.limit > 0 {
		args = append(args, q.limit)
		query += fmt.Sprintf(" ORDER BY v.updated_at DESC, v.mailbox_email DESC LIMIT $%d", len(args))
	}
	return query, args
}

func (q *virtualMailboxQuery) addArg(value any) string {
	q.args = append(q.args, value)
	return fmt.Sprintf("$%d", len(q.args))
}

func (q *virtualMailboxQuery) innerWhereSQL() string {
	if len(q.innerConditions) == 0 {
		return ""
	}
	return "AND " + strings.Join(q.innerConditions, " AND ")
}
