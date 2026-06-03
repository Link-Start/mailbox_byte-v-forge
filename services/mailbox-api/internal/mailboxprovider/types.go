package mailboxprovider

import (
	"context"

	"github.com/byte-v-forge/common-lib/pagex"
	"github.com/jackc/pgx/v5"

	"mailboxapi/internal/mailboxmodel"
)

type RuntimeContext interface {
	DomainsForProvider(provider string) []string
}

type SelectFields struct {
	Password     string
	RefreshToken string
	AccessToken  string
	AuthStatus   string
	LastError    string
}

type MailboxRecord struct {
	Email        string
	Provider     string
	RefreshToken string
	AuthStatus   string
}

type TokenFields struct {
	Table              string
	EmailColumn        string
	RefreshTokenColumn string
	AccessTokenColumn  string
	AuthStatusColumn   string
	LastErrorColumn    string
	UpdatedAtColumn    string
}

func (f TokenFields) HasTokenStorage() bool {
	return f.Table != "" && f.EmailColumn != "" && f.RefreshTokenColumn != "" && f.AccessTokenColumn != ""
}

type InboxRetention struct {
	TouchedMailboxes map[string]struct{}
	TouchedDomains   map[string]struct{}
}

type ListQuery struct {
	AuthStatus   string
	Provider     string
	EmailAddress string
	Cursor       pagex.KeysetCursor
	Limit        int
}

func (q ListQuery) ScanLimit() int {
	return pagex.KeysetLookaheadLimit(q.Limit)
}

func (q ListQuery) HasCursor() bool {
	return pagex.HasKeysetCursor(q.Cursor)
}

type UpsertFunc func(context.Context, pgx.Tx, *mailboxmodel.Record, int64) error
type AuthFilterFunc func(string, *[]any) string
type ValidatePollFunc func(MailboxRecord) error
type UpdateAuthFunc func(context.Context, pgx.Tx, string, string, string, int64) error
type PruneInboundFunc func(context.Context, pgx.Tx, InboxRetention) error
