package mailboxpg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"
)

type inboxMessageQuery struct {
	conditions []string
	args       []any
	orderBy    string
	limit      int
}

func newInboxMessageQuery() *inboxMessageQuery {
	return &inboxMessageQuery{}
}

func (q *inboxMessageQuery) WhereMailbox(email string) *inboxMessageQuery {
	q.conditions = append(q.conditions, fmt.Sprintf("mailbox_email = %s", q.addArg(emailx.Normalize(email))))
	return q
}

func (q *inboxMessageQuery) WhereReceivedAfter(receivedAt int64) *inboxMessageQuery {
	if receivedAt > 0 {
		q.conditions = append(q.conditions, fmt.Sprintf("received_at > %s", q.addArg(receivedAt)))
	}
	return q
}

func (q *inboxMessageQuery) WhereReceivedAtOrAfter(receivedAt int64) *inboxMessageQuery {
	if receivedAt > 0 {
		q.conditions = append(q.conditions, fmt.Sprintf("received_at >= %s", q.addArg(receivedAt)))
	}
	return q
}

func (q *inboxMessageQuery) WhereKeyword(keyword string) *inboxMessageQuery {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return q
	}
	placeholder := q.addArg("%" + keyword + "%")
	q.conditions = append(q.conditions, fmt.Sprintf("(subject ILIKE %s OR body_preview ILIKE %s OR body_text ILIKE %s)", placeholder, placeholder, placeholder))
	return q
}

func (q *inboxMessageQuery) OrderByLatest() *inboxMessageQuery {
	q.orderBy = "received_at DESC, updated_at DESC, message_key DESC"
	return q
}

func (q *inboxMessageQuery) Limit(limit int) *inboxMessageQuery {
	q.limit = limit
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
	return query, args
}

func (q *inboxMessageQuery) addArg(value any) string {
	q.args = append(q.args, value)
	return fmt.Sprintf("$%d", len(q.args))
}
