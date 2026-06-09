package mailboxpg

import (
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

func (m *MailboxRow) ToRecord(normalizeProvider func(string) string, prepareProjection func(*mailboxmodel.Record)) *mailboxmodel.Record {
	if m == nil {
		return nil
	}
	mailbox := &mailboxmodel.Record{
		EmailAddress: m.Email,
		ProviderKey:  normalizeProvider(m.Provider),
		Password:     m.Password,
		RefreshToken: m.RefreshToken,
		AccessToken:  m.AccessToken,
		AuthStatus:   m.AuthStatus,
		LastError:    m.LastError,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		Domain:       DomainForEmail(m.Email),
	}
	prepareProjection(mailbox)
	return mailbox
}

func (m *MailboxRow) ToProviderRecord() mailboxprovider.MailboxRecord {
	if m == nil {
		return mailboxprovider.MailboxRecord{}
	}
	return mailboxprovider.MailboxRecord{
		Email:        m.Email,
		Provider:     m.Provider,
		RefreshToken: m.RefreshToken,
		AuthStatus:   m.AuthStatus,
	}
}
