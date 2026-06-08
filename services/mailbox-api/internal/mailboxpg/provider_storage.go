package mailboxpg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"

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
