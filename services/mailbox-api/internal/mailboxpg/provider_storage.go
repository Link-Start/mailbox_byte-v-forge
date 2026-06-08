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
