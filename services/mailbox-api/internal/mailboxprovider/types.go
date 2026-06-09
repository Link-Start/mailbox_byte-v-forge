package mailboxprovider

type RuntimeContext interface {
	DomainsForProvider(provider string) []string
}

type MailboxRecord struct {
	Email        string
	Provider     string
	RefreshToken string
	AuthStatus   string
}

type ValidatePollFunc func(MailboxRecord) error
