package mailboxprovider

import (
	"fmt"
	"strings"
)

func (r *Registry) SchemaStatements() []string {
	statements := []string{}
	for _, provider := range r.StorageExtensions() {
		statements = append(statements, provider.SchemaStatements()...)
	}
	return statements
}

func (r *Registry) LegacyStatements() []string {
	statements := []string{}
	for _, provider := range r.StorageExtensions() {
		statements = append(statements, provider.PrepareLegacyData()...)
	}
	return statements
}

func (r *Registry) MailboxSelectSQL() string {
	fields := r.FieldExpressions()
	joins := ""
	for _, provider := range r.StorageExtensions() {
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

func (r *Registry) FieldExpressions() SelectFields {
	expressions := SelectFields{
		Password:     "''",
		RefreshToken: "''",
		AccessToken:  "''",
		AuthStatus:   "''",
		LastError:    "''",
	}
	for _, provider := range r.StorageExtensions() {
		fields := provider.SelectFields()
		expressions.Password = coalesceField(expressions.Password, fields.Password)
		expressions.RefreshToken = coalesceField(expressions.RefreshToken, fields.RefreshToken)
		expressions.AccessToken = coalesceField(expressions.AccessToken, fields.AccessToken)
		expressions.AuthStatus = coalesceField(expressions.AuthStatus, fields.AuthStatus)
		expressions.LastError = coalesceField(expressions.LastError, fields.LastError)
	}
	return expressions
}

func coalesceField(current string, next string) string {
	next = strings.TrimSpace(next)
	if next == "" {
		return current
	}
	if current == "''" {
		return fmt.Sprintf("COALESCE(%s, '')", next)
	}
	return fmt.Sprintf("COALESCE(NULLIF(%s, ''), %s)", next, current)
}
