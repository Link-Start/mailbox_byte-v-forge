package mailboxpg

import (
	"context"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/jackc/pgx/v5"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

func (r *Repository) upsertProviderMailboxData(ctx context.Context, tx pgx.Tx, provider string, mailbox *mailboxmodel.Record, now int64) error {
	definition := r.providers.StorageByKey(provider)
	if definition == nil {
		return nil
	}
	fields, ok := definition.TokenFields()
	if !ok {
		return nil
	}
	authStatus := strings.TrimSpace(mailbox.GetAuthStatus())
	explicitAuthStatus := authStatus
	if authStatus == "" {
		authStatus = mailboxmodel.AuthStatusOAuthPending
		if strings.TrimSpace(mailbox.GetRefreshToken()) != "" {
			authStatus = mailboxmodel.AuthStatusAuthorized
		}
	}

	builder, err := newProviderStorageBuilder(fields)
	if err != nil {
		return err
	}
	builder.add(fields.EmailColumn, emailx.Normalize(mailbox.GetEmailAddress()), "")
	builder.add(fields.PasswordColumn, strings.TrimSpace(mailbox.GetPassword()), nonEmptyExcludedUpdate)
	builder.add(fields.RefreshTokenColumn, strings.TrimSpace(mailbox.GetRefreshToken()), nonEmptyExcludedUpdate)
	builder.add(fields.AccessTokenColumn, strings.TrimSpace(mailbox.GetAccessToken()), nonEmptyExcludedUpdate)
	authColumn := builder.add(fields.AuthStatusColumn, authStatus, "")
	lastErrorColumn := builder.add(fields.LastErrorColumn, strings.TrimSpace(mailbox.GetLastError()), "")
	builder.add(fields.CreatedAtColumn, now, "")
	builder.add(fields.UpdatedAtColumn, now, excludedUpdate)

	extraArgs := []any{}
	if authColumn != "" {
		extraArgs = append(extraArgs, explicitAuthStatus)
		explicitAuthArg := len(builder.args) + len(extraArgs)
		refreshColumn, err := mailboxprovider.SQLIdentifier(fields.RefreshTokenColumn)
		if err != nil {
			return err
		}
		builder.setUpdate(authColumn, fmt.Sprintf(
			"%s = CASE WHEN $%d <> '' THEN EXCLUDED.%s WHEN EXCLUDED.%s <> '' THEN '%s' ELSE %s.%s END",
			authColumn, explicitAuthArg, authColumn, refreshColumn, mailboxmodel.AuthStatusAuthorized, builder.table, authColumn,
		))
		if lastErrorColumn != "" {
			builder.setUpdate(lastErrorColumn, fmt.Sprintf(
				"%s = CASE WHEN $%d <> '' OR EXCLUDED.%s <> '' THEN EXCLUDED.%s ELSE %s.%s END",
				lastErrorColumn, explicitAuthArg, lastErrorColumn, lastErrorColumn, builder.table, lastErrorColumn,
			))
		}
	}
	return builder.exec(ctx, tx, extraArgs...)
}

func (r *Repository) updateProviderAuth(ctx context.Context, tx pgx.Tx, provider string, email string, authStatus string, lastError string, now int64) error {
	definition := r.providers.StorageByKey(provider)
	if definition == nil {
		return fmt.Errorf("mailbox provider has no auth state: %s", provider)
	}
	fields, ok := definition.TokenFields()
	if !ok || fields.AuthStatusColumn == "" {
		return fmt.Errorf("mailbox provider has no auth state: %s", provider)
	}
	table, err := mailboxprovider.SQLIdentifier(fields.Table)
	if err != nil {
		return err
	}
	emailColumn, err := mailboxprovider.SQLIdentifier(fields.EmailColumn)
	if err != nil {
		return err
	}
	authColumn, err := mailboxprovider.SQLIdentifier(fields.AuthStatusColumn)
	if err != nil {
		return err
	}
	args := []any{strings.TrimSpace(authStatus)}
	assignments := []string{fmt.Sprintf("%s = $1", authColumn)}
	if fields.LastErrorColumn != "" {
		column, err := mailboxprovider.SQLIdentifier(fields.LastErrorColumn)
		if err != nil {
			return err
		}
		args = append(args, strings.TrimSpace(lastError))
		assignments = append(assignments, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	if fields.UpdatedAtColumn != "" {
		column, err := mailboxprovider.SQLIdentifier(fields.UpdatedAtColumn)
		if err != nil {
			return err
		}
		args = append(args, now)
		assignments = append(assignments, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	args = append(args, emailx.Normalize(email))
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d", table, strings.Join(assignments, ", "), emailColumn, len(args))
	_, err = tx.Exec(ctx, query, args...)
	return err
}

type providerStorageBuilder struct {
	table       string
	emailColumn string
	columns     []string
	args        []any
	updates     map[string]string
	err         error
}

const (
	excludedUpdate         = "excluded"
	nonEmptyExcludedUpdate = "non_empty_excluded"
)

func newProviderStorageBuilder(fields mailboxprovider.TokenFields) (*providerStorageBuilder, error) {
	table, err := mailboxprovider.SQLIdentifier(fields.Table)
	if err != nil {
		return nil, err
	}
	emailColumn, err := mailboxprovider.SQLIdentifier(fields.EmailColumn)
	if err != nil {
		return nil, err
	}
	return &providerStorageBuilder{table: table, emailColumn: emailColumn, updates: map[string]string{}}, nil
}

func (b *providerStorageBuilder) add(rawColumn string, value any, updateMode string) string {
	if strings.TrimSpace(rawColumn) == "" {
		return ""
	}
	column, err := mailboxprovider.SQLIdentifier(rawColumn)
	if err != nil {
		b.err = err
		return ""
	}
	b.columns = append(b.columns, column)
	b.args = append(b.args, value)
	switch updateMode {
	case excludedUpdate:
		b.updates[column] = fmt.Sprintf("%s = EXCLUDED.%s", column, column)
	case nonEmptyExcludedUpdate:
		b.updates[column] = fmt.Sprintf("%s = CASE WHEN EXCLUDED.%s <> '' THEN EXCLUDED.%s ELSE %s.%s END", column, column, column, b.table, column)
	}
	return column
}

func (b *providerStorageBuilder) setUpdate(column string, expression string) {
	if column != "" && strings.TrimSpace(expression) != "" {
		b.updates[column] = expression
	}
}

func (b *providerStorageBuilder) exec(ctx context.Context, tx pgx.Tx, extraArgs ...any) error {
	if b.err != nil {
		return b.err
	}
	placeholders := make([]string, 0, len(b.args))
	for index := range b.args {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
	}
	updates := make([]string, 0, len(b.updates))
	for _, column := range b.columns {
		if update := b.updates[column]; update != "" {
			updates = append(updates, update)
		}
	}
	args := append([]any{}, b.args...)
	args = append(args, extraArgs...)
	if len(updates) == 0 {
		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO NOTHING",
			b.table,
			strings.Join(b.columns, ", "),
			strings.Join(placeholders, ", "),
			b.emailColumn,
		)
		_, err := tx.Exec(ctx, query, args...)
		return err
	}
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		b.table,
		strings.Join(b.columns, ", "),
		strings.Join(placeholders, ", "),
		b.emailColumn,
		strings.Join(updates, ", "),
	)
	_, err := tx.Exec(ctx, query, args...)
	return err
}
