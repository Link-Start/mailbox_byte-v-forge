package main

import (
	"fmt"
	"strings"

	"mailboxapi/internal/mailboxprovider"
)

func (s *MailboxStore) mailboxProviderSchemaStatements() []string {
	statements := []string{}
	for _, provider := range s.mailboxProviderStorageExtensions() {
		statements = append(statements, provider.SchemaStatements()...)
	}
	return statements
}

func (s *MailboxStore) mailboxProviderLegacyStatements() []string {
	statements := []string{}
	for _, provider := range s.mailboxProviderStorageExtensions() {
		statements = append(statements, provider.PrepareLegacyData()...)
	}
	return statements
}

func (s *MailboxStore) mailboxSelectSQL() string {
	fields := s.mailboxProviderFieldExpressions()
	joins := ""
	for _, provider := range s.mailboxProviderStorageExtensions() {
		if join := strings.TrimSpace(provider.SelectJoin()); join != "" {
			joins += "\n" + join
		}
	}
	return fmt.Sprintf(`
	SELECT m.id, m.email, m.provider,
		%s AS password,
		%s AS refresh_token,
		%s AS access_token,
		%s AS auth_status,
		%s AS last_error,
		m.created_at, m.updated_at
	FROM mailboxes m%s
`, fields.Password, fields.RefreshToken, fields.AccessToken, fields.AuthStatus, fields.LastError, joins)
}

func (s *MailboxStore) mailboxProviderFieldExpressions() mailboxprovider.SelectFields {
	expressions := mailboxprovider.SelectFields{
		Password:     "''",
		RefreshToken: "''",
		AccessToken:  "''",
		AuthStatus:   "''",
		LastError:    "''",
	}
	for _, provider := range s.mailboxProviderStorageExtensions() {
		fields := provider.SelectFields()
		expressions.Password = coalesceProviderField(expressions.Password, fields.Password)
		expressions.RefreshToken = coalesceProviderField(expressions.RefreshToken, fields.RefreshToken)
		expressions.AccessToken = coalesceProviderField(expressions.AccessToken, fields.AccessToken)
		expressions.AuthStatus = coalesceProviderField(expressions.AuthStatus, fields.AuthStatus)
		expressions.LastError = coalesceProviderField(expressions.LastError, fields.LastError)
	}
	return expressions
}

func coalesceProviderField(current string, next string) string {
	next = strings.TrimSpace(next)
	if next == "" {
		return current
	}
	if current == "''" {
		return fmt.Sprintf("COALESCE(%s, '')", next)
	}
	return fmt.Sprintf("COALESCE(NULLIF(%s, ''), %s)", next, current)
}
