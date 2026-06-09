package mailboxpg

import (
	"fmt"
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/pagex"
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
