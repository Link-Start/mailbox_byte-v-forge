package mailboxmodel

import mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

type Record struct {
	EmailAddress string
	Password     string
	RefreshToken string
	AccessToken  string
	LastError    string
	CreatedAt    int64
	UpdatedAt    int64
	AuthStatus   string
	ProviderKey  string
	LatestSignal *mailboxv1.EmailSignal
	Domain       string
}

func (m *Record) GetEmailAddress() string {
	if m == nil {
		return ""
	}
	return m.EmailAddress
}

func (m *Record) GetProviderKey() string {
	if m == nil {
		return ""
	}
	return m.ProviderKey
}

func (m *Record) GetDomain() string {
	if m == nil {
		return ""
	}
	return m.Domain
}
