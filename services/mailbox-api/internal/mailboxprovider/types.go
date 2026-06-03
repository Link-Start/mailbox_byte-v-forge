package mailboxprovider

import (
	"context"

	"github.com/byte-v-forge/common-lib/pagex"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mailboxapi/pb"
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

type UpsertFunc func(context.Context, pgx.Tx, *pb.EmailMailbox, int64) error
type AuthFilterFunc func(string, *[]any) string
type ValidatePollFunc func(MailboxRecord) error
type UpdateAuthFunc func(context.Context, pgx.Tx, string, string, string, int64) error
type UpdateTokensFunc func(context.Context, *pgxpool.Pool, string, string, string) error
type PruneInboundFunc func(context.Context, pgx.Tx, InboxRetention) error
type VirtualMailboxesFunc func(context.Context, *pgxpool.Pool, ListQuery) ([]*pb.EmailMailbox, error)
