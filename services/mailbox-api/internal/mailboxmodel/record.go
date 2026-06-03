package mailboxmodel

import mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

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

func (m *Record) GetPassword() string {
	if m == nil {
		return ""
	}
	return m.Password
}

func (m *Record) GetRefreshToken() string {
	if m == nil {
		return ""
	}
	return m.RefreshToken
}

func (m *Record) GetAccessToken() string {
	if m == nil {
		return ""
	}
	return m.AccessToken
}

func (m *Record) GetLastError() string {
	if m == nil {
		return ""
	}
	return m.LastError
}

func (m *Record) GetCreatedAt() int64 {
	if m == nil {
		return 0
	}
	return m.CreatedAt
}

func (m *Record) GetUpdatedAt() int64 {
	if m == nil {
		return 0
	}
	return m.UpdatedAt
}

func (m *Record) GetAuthStatus() string {
	if m == nil {
		return ""
	}
	return m.AuthStatus
}

func (m *Record) GetProviderKey() string {
	if m == nil {
		return ""
	}
	return m.ProviderKey
}

func (m *Record) GetLatestSignal() *mailboxv1.EmailSignal {
	if m == nil {
		return nil
	}
	return m.LatestSignal
}

func (m *Record) GetDomain() string {
	if m == nil {
		return ""
	}
	return m.Domain
}
