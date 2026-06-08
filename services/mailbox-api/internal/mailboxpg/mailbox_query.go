package mailboxpg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/pagex"

	"mailboxapi/internal/mailboxprovider"
)

type mailboxSelectQuery struct {
	providers  *mailboxprovider.Registry
	conditions []string
	args       []any
	orderBy    string
	limit      int
	err        error
}

type mailboxQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type mailboxSelectFields struct {
	Password     string
	RefreshToken string
	AccessToken  string
	AuthStatus   string
	LastError    string
}

func (r *Repository) newMailboxSelectQuery() *mailboxSelectQuery {
	return &mailboxSelectQuery{providers: r.providers}
}

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

func lockStoredMailbox(ctx context.Context, tx pgx.Tx, email string) (string, error) {
	var provider string
	err := tx.QueryRow(ctx, "SELECT provider FROM mailboxes WHERE email = $1 FOR UPDATE", emailx.Normalize(email)).Scan(&provider)
	if err != nil {
		return "", err
	}
	return provider, nil
}
