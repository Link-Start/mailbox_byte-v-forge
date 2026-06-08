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
	err        error
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
	condition, ok, err := q.providerAuthFilter(provider, authStatus)
	if err != nil {
		q.err = err
		return q
	}
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
	query, args, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return querier.Query(ctx, query, args...)
}

func (q *mailboxSelectQuery) ScanOne(ctx context.Context, querier mailboxQuerier) (*MailboxRow, error) {
	query, args, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return ScanMailbox(querier.QueryRow(ctx, query, args...))
}

func (q *mailboxSelectQuery) SQL() (string, []any, error) {
	if q.err != nil {
		return "", nil, q.err
	}
	query, err := q.selectSQL()
	if err != nil {
		return "", nil, err
	}
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
	return query, args, nil
}

func (q *mailboxSelectQuery) addArg(value any) string {
	q.args = append(q.args, value)
	return fmt.Sprintf("$%d", len(q.args))
}

func (q *mailboxSelectQuery) selectSQL() (string, error) {
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
		join, err := storageJoinSQL(provider, tokenFields, alias)
		if err != nil {
			return "", err
		}
		joins = append(joins, join)
		if err := q.mergeStorageFields(&fields, tokenFields, alias); err != nil {
			return "", err
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
`, fields.Password, fields.RefreshToken, fields.AccessToken, fields.AuthStatus, fields.LastError, strings.Join(joins, "")), nil
}

func (q *mailboxSelectQuery) mergeStorageFields(fields *mailboxSelectFields, tokenFields mailboxprovider.TokenFields, alias string) error {
	var err error
	if fields.Password, err = coalesceStorageField(fields.Password, alias, tokenFields.PasswordColumn); err != nil {
		return err
	}
	if fields.RefreshToken, err = coalesceStorageField(fields.RefreshToken, alias, tokenFields.RefreshTokenColumn); err != nil {
		return err
	}
	if fields.AccessToken, err = coalesceStorageField(fields.AccessToken, alias, tokenFields.AccessTokenColumn); err != nil {
		return err
	}
	if fields.AuthStatus, err = coalesceStorageField(fields.AuthStatus, alias, tokenFields.AuthStatusColumn); err != nil {
		return err
	}
	if fields.LastError, err = coalesceStorageField(fields.LastError, alias, tokenFields.LastErrorColumn); err != nil {
		return err
	}
	return nil
}

func (q *mailboxSelectQuery) providerAuthFilter(provider string, authStatus string) (string, bool, error) {
	definition := q.providers.StorageByKey(provider)
	if definition != nil {
		return q.providerAuthFilterSQL(definition, authStatus)
	}
	parts := []string{}
	for _, definition := range q.providers.StorageExtensions() {
		filter, ok, err := q.providerAuthFilterSQL(definition, authStatus)
		if err != nil {
			return "", false, err
		}
		if ok {
			parts = append(parts, filter)
		}
	}
	if len(parts) == 0 {
		return "", false, nil
	}
	return "(" + strings.Join(parts, " OR ") + ")", true, nil
}

func (q *mailboxSelectQuery) providerAuthFilterSQL(provider mailboxprovider.StorageExtension, authStatus string) (string, bool, error) {
	fields, ok := provider.TokenFields()
	if !ok || strings.TrimSpace(fields.AuthStatusColumn) == "" {
		return "", false, nil
	}
	column, err := sqlIdentifier(fields.AuthStatusColumn)
	if err != nil {
		return "", false, fmt.Errorf("provider %s auth status column: %w", provider.Key(), err)
	}
	return fmt.Sprintf("%s.%s = %s", providerStorageAlias(provider.Key()), column, q.addArg(strings.TrimSpace(authStatus))), true, nil
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

func storageJoinSQL(provider mailboxprovider.StorageExtension, fields mailboxprovider.TokenFields, alias string) (string, error) {
	table, err := sqlIdentifier(fields.Table)
	if err != nil {
		return "", fmt.Errorf("provider %s storage table: %w", provider.Key(), err)
	}
	emailColumn, err := sqlIdentifier(fields.EmailColumn)
	if err != nil {
		return "", fmt.Errorf("provider %s email column: %w", provider.Key(), err)
	}
	return fmt.Sprintf(
		"\nLEFT JOIN %s %s ON %s.%s = m.email AND m.provider = %s",
		table,
		alias,
		alias,
		emailColumn,
		sqlStringLiteral(provider.Key()),
	), nil
}

func coalesceStorageField(current string, alias string, column string) (string, error) {
	next, err := storageColumnSQL(alias, column)
	if err != nil {
		return current, err
	}
	return coalesceField(current, next), nil
}

func storageColumnSQL(alias string, column string) (string, error) {
	if strings.TrimSpace(column) == "" {
		return "", nil
	}
	identifier, err := sqlIdentifier(column)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s", alias, identifier), nil
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

func sqlIdentifier(value string) (string, error) {
	return mailboxprovider.SQLIdentifier(value)
}

func sqlStringLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
