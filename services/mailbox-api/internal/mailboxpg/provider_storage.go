package mailboxpg

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"

	"mailboxapi/internal/mailboxmodel"
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
	authStatus, explicitAuthStatus := providerMailboxAuthStatus(mailbox)
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

	extraArgs, err := providerStorageAuthUpdates(builder, fields, authColumn, lastErrorColumn, explicitAuthStatus)
	if err != nil {
		return err
	}
	return builder.exec(ctx, tx, extraArgs...)
}
