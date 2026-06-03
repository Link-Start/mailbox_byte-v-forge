package main

import (
	"fmt"
	"strings"

	"mailboxapi/internal/mailboxprovider"
)

func mailboxProviderSchemaStatements() []string {
	statements := []string{}
	for _, provider := range mailboxProviderStorageExtensions() {
		statements = append(statements, provider.SchemaStatements()...)
	}
	return statements
}

func mailboxProviderLegacyStatements() []string {
	statements := []string{}
	for _, provider := range mailboxProviderStorageExtensions() {
		statements = append(statements, provider.PrepareLegacyData()...)
	}
	return statements
}

func mailboxSelectSQL() string {
	fields := mailboxProviderFieldExpressions()
	joins := ""
	for _, provider := range mailboxProviderStorageExtensions() {
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

func mailboxProviderFieldExpressions() mailboxprovider.SelectFields {
	expressions := mailboxprovider.SelectFields{
		Password:     "''",
		RefreshToken: "''",
		AccessToken:  "''",
		AuthStatus:   "''",
		LastError:    "''",
	}
	for _, provider := range mailboxProviderStorageExtensions() {
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
