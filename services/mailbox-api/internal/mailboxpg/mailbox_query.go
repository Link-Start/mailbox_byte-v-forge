package mailboxpg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/pagex"

	"mailboxapi/internal/mailboxprovider"
)

type mailboxSelectQuery struct {
	providers  *mailboxprovider.Registry
	conditions []string
	args       []any
	orderBy    string
	limit      int
}

type mailboxQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type mailboxSelectFields struct {
	Password     string
	RefreshToken string
	AccessToken  string
	AuthStatus   string
	LastError    string
}

func (r *Repository) newMailboxSelectQuery() *mailboxSelectQuery {
	return &mailboxSelectQuery{providers: r.providers}
}

func (q *mailboxSelectQuery) WhereEmail(email string) *mailboxSelectQuery {
	q.conditions = append(q.conditions, fmt.Sprintf("m.email = %s", q.addArg(emailx.Normalize(email))))
	return q
}

func (q *mailboxSelectQuery) WhereProvider(provider string) *mailboxSelectQuery {
	q.conditions = append(q.conditions, fmt.Sprintf("m.provider = %s", q.addArg(provider)))
	return q
}

func (q *mailboxSelectQuery) WhereAuthStatus(provider string, authStatus string) *mailboxSelectQuery {
	condition, ok := q.providerAuthFilter(provider, authStatus)
	if !ok {
		condition = "FALSE"
	}
	q.conditions = append(q.conditions, condition)
	return q
}

func (q *mailboxSelectQuery) WhereCursor(cursor pagex.KeysetCursor) *mailboxSelectQuery {
	updatedAtParam := q.addArg(cursor.UpdatedAt.Unix())
	emailParam := q.addArg(emailx.Normalize(cursor.ID))
	q.conditions = append(q.conditions, fmt.Sprintf("(m.updated_at < %s OR (m.updated_at = %s AND m.email < %s))", updatedAtParam, updatedAtParam, emailParam))
	return q
}

func (q *mailboxSelectQuery) OrderByUpdatedDesc() *mailboxSelectQuery {
	q.orderBy = "m.updated_at DESC, m.email DESC"
	return q
}

func (q *mailboxSelectQuery) Limit(limit int) *mailboxSelectQuery {
	q.limit = limit
	return q
}

func (q *mailboxSelectQuery) Query(ctx context.Context, querier mailboxQuerier) (pgx.Rows, error) {
	query, args := q.SQL()
	return querier.Query(ctx, query, args...)
}

func (q *mailboxSelectQuery) ScanOne(ctx context.Context, querier mailboxQuerier) (*MailboxRow, error) {
	query, args := q.SQL()
	return ScanMailbox(querier.QueryRow(ctx, query, args...))
}

func (q *mailboxSelectQuery) SQL() (string, []any) {
	query := q.selectSQL()
	if len(q.conditions) > 0 {
		query += " WHERE " + strings.Join(q.conditions, " AND ")
	}
	if strings.TrimSpace(q.orderBy) != "" {
		query += " ORDER BY " + q.orderBy
	}
	args := append([]any{}, q.args...)
	if q.limit > 0 {
		args = append(args, q.limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	return query, args
}

func (q *mailboxSelectQuery) addArg(value any) string {
	q.args = append(q.args, value)
	return fmt.Sprintf("$%d", len(q.args))
}

func (q *mailboxSelectQuery) selectSQL() string {
	fields := mailboxSelectFields{
		Password:     "''",
		RefreshToken: "''",
		AccessToken:  "''",
		AuthStatus:   "''",
		LastError:    "''",
	}
	joins := []string{}
	for _, provider := range q.providers.StorageExtensions() {
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

func (q *mailboxSelectQuery) providerAuthFilter(provider string, authStatus string) (string, bool) {
	definition := q.providers.StorageByKey(provider)
	if definition != nil {
		return q.providerAuthFilterSQL(definition, authStatus)
	}
	parts := []string{}
	for _, definition := range q.providers.StorageExtensions() {
		if filter, ok := q.providerAuthFilterSQL(definition, authStatus); ok {
			parts = append(parts, filter)
		}
	}
	if len(parts) == 0 {
		return "", false
	}
	return "(" + strings.Join(parts, " OR ") + ")", true
}

func (q *mailboxSelectQuery) providerAuthFilterSQL(provider mailboxprovider.StorageExtension, authStatus string) (string, bool) {
	fields, ok := provider.TokenFields()
	if !ok || strings.TrimSpace(fields.AuthStatusColumn) == "" {
		return "", false
	}
	return fmt.Sprintf("%s.%s = %s", providerStorageAlias(provider.Key()), mustSQLIdentifier(fields.AuthStatusColumn), q.addArg(strings.TrimSpace(authStatus))), true
}

func lockStoredMailbox(ctx context.Context, tx pgx.Tx, email string) (string, error) {
	var provider string
	err := tx.QueryRow(ctx, "SELECT provider FROM mailboxes WHERE email = $1 FOR UPDATE", emailx.Normalize(email)).Scan(&provider)
	if err != nil {
		return "", err
	}
	return provider, nil
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

func storageJoinSQL(provider mailboxprovider.StorageExtension, fields mailboxprovider.TokenFields, alias string) string {
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
	provider = mailboxprovider.NormalizeKey(provider)
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
	identifier, err := mailboxprovider.SQLIdentifier(value)
	if err != nil {
		panic(err)
	}
	return identifier
}

func sqlStringLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
