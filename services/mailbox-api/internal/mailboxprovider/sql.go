package mailboxprovider

import (
	"fmt"
	"strings"
)

type mailboxSelectFields struct {
	Password     string
	RefreshToken string
	AccessToken  string
	AuthStatus   string
	LastError    string
}

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
	fields := mailboxSelectFields{
		Password:     "''",
		RefreshToken: "''",
		AccessToken:  "''",
		AuthStatus:   "''",
		LastError:    "''",
	}
	joins := []string{}
	for _, provider := range r.StorageExtensions() {
		tokenFields, ok := provider.TokenFields()
		if !ok {
			continue
		}
		alias := providerStorageAlias(provider.Key())
		joins = append(joins, storageJoinSQL(provider, tokenFields, alias))
		fields.Password = coalesceField(fields.Password, storageColumnSQL(alias, tokenFields.PasswordColumn))
		fields.RefreshToken = coalesceField(fields.RefreshToken, storageColumnSQL(alias, tokenFields.RefreshTokenColumn))
		fields.AccessToken = coalesceField(fields.AccessToken, storageColumnSQL(alias, tokenFields.AccessTokenColumn))
		fields.AuthStatus = coalesceField(fields.AuthStatus, storageColumnSQL(alias, tokenFields.AuthStatusColumn))
		fields.LastError = coalesceField(fields.LastError, storageColumnSQL(alias, tokenFields.LastErrorColumn))
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
`, fields.Password, fields.RefreshToken, fields.AccessToken, fields.AuthStatus, fields.LastError, strings.Join(joins, ""))
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

func providerAuthFilterSQL(provider StorageExtension, authStatus string, args *[]any) (string, bool) {
	fields, ok := provider.TokenFields()
	if !ok || strings.TrimSpace(fields.AuthStatusColumn) == "" {
		return "", false
	}
	argsValue := strings.TrimSpace(authStatus)
	*args = append(*args, argsValue)
	return fmt.Sprintf("%s.%s = $%d", providerStorageAlias(provider.Key()), mustSQLIdentifier(fields.AuthStatusColumn), len(*args)), true
}

func storageJoinSQL(provider StorageExtension, fields TokenFields, alias string) string {
	return fmt.Sprintf(
		"\nLEFT JOIN %s %s ON %s.%s = m.email AND m.provider = %s",
		mustSQLIdentifier(fields.Table),
		alias,
		alias,
		mustSQLIdentifier(fields.EmailColumn),
		sqlStringLiteral(provider.Key()),
	)
}

func storageColumnSQL(alias string, column string) string {
	if strings.TrimSpace(column) == "" {
		return ""
	}
	return fmt.Sprintf("%s.%s", alias, mustSQLIdentifier(column))
}

func providerStorageAlias(provider string) string {
	provider = NormalizeKey(provider)
	var out strings.Builder
	out.WriteString("provider")
	wroteProviderKey := false
	wroteSeparator := false
	for _, r := range provider {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			if !wroteProviderKey {
				out.WriteRune('_')
				wroteProviderKey = true
			}
			out.WriteRune(r)
			wroteSeparator = false
		default:
			if wroteProviderKey && !wroteSeparator {
				out.WriteRune('_')
				wroteSeparator = true
			}
		}
	}
	return strings.TrimRight(out.String(), "_")
}

func mustSQLIdentifier(value string) string {
	identifier, err := SQLIdentifier(value)
	if err != nil {
		panic(err)
	}
	return identifier
}

func sqlStringLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
