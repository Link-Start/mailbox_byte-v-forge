package mailboxpg

import (
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

type Scanner interface {
	Scan(dest ...any) error
}

type MailboxRow struct {
	ID           string
	Email        string
	Provider     string
	Password     string
	RefreshToken string
	AccessToken  string
	AuthStatus   string
	LastError    string
	CreatedAt    int64
	UpdatedAt    int64
}

func ScanMailbox(scanner Scanner) (*MailboxRow, error) {
	var row MailboxRow
	err := scanner.Scan(
		&row.ID,
		&row.Email,
		&row.Provider,
		&row.Password,
		&row.RefreshToken,
		&row.AccessToken,
		&row.AuthStatus,
		&row.LastError,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

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

func DomainForEmail(email string) string {
	_, domain, ok := strings.Cut(emailx.Normalize(email), "@")
	if !ok {
		return ""
	}
	return domain
}
