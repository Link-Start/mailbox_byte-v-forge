package mailboxpg

import (
	"context"

	"github.com/jackc/pgx/v5"
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
