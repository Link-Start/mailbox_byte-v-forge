package main

import (
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"

	"mailboxapi/pb"
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

func (m *mailboxRow) toProto() *pb.EmailMailbox {
	if m == nil {
		return nil
	}
	mailbox := &pb.EmailMailbox{
		EmailAddress: m.Email,
		ProviderKey:  normalizeEmailProvider(m.Provider),
		Password:     m.Password,
		RefreshToken: m.RefreshToken,
		AccessToken:  m.AccessToken,
		AuthStatus:   m.AuthStatus,
		LastError:    m.LastError,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		Domain:       domainForEmail(m.Email),
	}
	prepareMailboxProjection(mailbox)
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
