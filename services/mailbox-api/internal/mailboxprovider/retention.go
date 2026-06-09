package mailboxprovider

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
