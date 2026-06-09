package mailboxpg

import (
	"fmt"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/pagex"
)

func (q *mailboxSelectQuery) WhereEmail(email string) *mailboxSelectQuery {
	q.conditions = append(q.conditions, fmt.Sprintf("m.email = %s", q.addArg(emailx.Normalize(email))))
	return q
}

func (q *mailboxSelectQuery) WhereProvider(provider string) *mailboxSelectQuery {
	q.conditions = append(q.conditions, fmt.Sprintf("m.provider = %s", q.addArg(provider)))
	return q
}

func (q *mailboxSelectQuery) WhereAuthStatus(provider string, authStatus string) *mailboxSelectQuery {
	condition, ok, err := q.providerAuthFilter(provider, authStatus)
	if err != nil {
		q.err = err
		return q
	}
	if !ok {
		condition = "FALSE"
	}
	q.conditions = append(q.conditions, condition)
	return q
}

func (q *mailboxSelectQuery) WhereCursor(cursor pagex.KeysetCursor) *mailboxSelectQuery {
	updatedAtParam := q.addArg(cursor.UpdatedAt.Unix())
	emailParam := q.addArg(emailx.Normalize(cursor.ID))
	q.conditions = append(q.conditions, fmt.Sprintf("(m.updated_at < %s OR (m.updated_at = %s AND m.email < %s))", updatedAtParam, updatedAtParam, emailParam))
	return q
}

func (q *mailboxSelectQuery) OrderByUpdatedDesc() *mailboxSelectQuery {
	q.orderBy = "m.updated_at DESC, m.email DESC"
	return q
}

func (q *mailboxSelectQuery) Limit(limit int) *mailboxSelectQuery {
	q.limit = limit
	return q
}
