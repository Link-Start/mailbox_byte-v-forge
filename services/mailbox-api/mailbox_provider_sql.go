package main

import (
	"fmt"
	"strings"
)

func mailboxProviderSchemaStatements() []string {
	statements := []string{}
	for _, provider := range mailboxProviderPlugins() {
		if provider.schemaStatements != nil {
			statements = append(statements, provider.schemaStatements()...)
		}
	}
	return statements
}

func mailboxProviderLegacyStatements() []string {
	statements := []string{}
	for _, provider := range mailboxProviderPlugins() {
		if provider.prepareLegacyData != nil {
			statements = append(statements, provider.prepareLegacyData()...)
		}
	}
	return statements
}

func mailboxSelectSQL() string {
	fields := mailboxProviderFieldExpressions()
	joins := ""
	for _, provider := range mailboxProviderPlugins() {
		if strings.TrimSpace(provider.selectJoin) != "" {
			joins += "\n" + provider.selectJoin
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
`, fields.password, fields.refreshToken, fields.accessToken, fields.authStatus, fields.lastError, joins)
}

func mailboxProviderFieldExpressions() mailboxProviderSelectFields {
	expressions := mailboxProviderSelectFields{
		password:     "''",
		refreshToken: "''",
		accessToken:  "''",
		authStatus:   "''",
		lastError:    "''",
	}
	for _, provider := range mailboxProviderPlugins() {
		fields := provider.selectFields
		expressions.password = coalesceProviderField(expressions.password, fields.password)
		expressions.refreshToken = coalesceProviderField(expressions.refreshToken, fields.refreshToken)
		expressions.accessToken = coalesceProviderField(expressions.accessToken, fields.accessToken)
		expressions.authStatus = coalesceProviderField(expressions.authStatus, fields.authStatus)
		expressions.lastError = coalesceProviderField(expressions.lastError, fields.lastError)
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
