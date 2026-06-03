package mailboxprovider

import "github.com/byte-v-forge/common-lib/pagex"

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
	PasswordColumn     string
	RefreshTokenColumn string
	AccessTokenColumn  string
	AuthStatusColumn   string
	LastErrorColumn    string
	CreatedAtColumn    string
	UpdatedAtColumn    string
}

func (f TokenFields) HasTokenStorage() bool {
	return f.Table != "" && f.EmailColumn != "" && f.RefreshTokenColumn != "" && f.AccessTokenColumn != ""
}

type RetentionScope string

const (
	RetentionScopeMailbox RetentionScope = "mailbox"
	RetentionScopeDomain  RetentionScope = "domain"
)

type MessageRetention struct {
	Scope       RetentionScope
	MaxMessages int
}

func (r MessageRetention) HasRetention() bool {
	return r.Scope != "" && r.MaxMessages > 0
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

type AuthFilterFunc func(string, *[]any) string
type ValidatePollFunc func(MailboxRecord) error
