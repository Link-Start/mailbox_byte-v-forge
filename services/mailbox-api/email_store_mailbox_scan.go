package main

import (
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"

	"mailboxapi/internal/mailboxmodel"
)

func scanMailbox(scanner rowScanner) (*mailboxRow, error) {
	var row mailboxRow
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

func (m *mailboxRow) toRecord() *mailboxmodel.Record {
	return mailboxRecordFromRow(m, normalizeEmailProvider, prepareMailboxProjection)
}

func (s *MailboxStore) recordFromMailboxRow(row *mailboxRow) *mailboxmodel.Record {
	return mailboxRecordFromRow(row, s.normalizeMailboxProviderInput, s.prepareMailboxProjection)
}

func mailboxRecordFromRow(row *mailboxRow, normalizeProvider func(string) string, prepareProjection func(*mailboxmodel.Record)) *mailboxmodel.Record {
	if row == nil {
		return nil
	}
	mailbox := &mailboxmodel.Record{
		EmailAddress: row.Email,
		ProviderKey:  normalizeProvider(row.Provider),
		Password:     row.Password,
		RefreshToken: row.RefreshToken,
		AccessToken:  row.AccessToken,
		AuthStatus:   row.AuthStatus,
		LastError:    row.LastError,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		Domain:       domainForEmail(row.Email),
	}
	prepareProjection(mailbox)
	return mailbox
}

func normalizeEmailProvider(provider string) string {
	return normalizeMailboxProviderInput(provider)
}

func domainForEmail(email string) string {
	_, domain, ok := strings.Cut(emailx.Normalize(email), "@")
	if !ok {
		return ""
	}
	return domain
}
