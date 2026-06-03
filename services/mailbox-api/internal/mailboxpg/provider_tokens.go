package mailboxpg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

func (r *Repository) updateMailboxTokens(ctx context.Context, provider string, email string, refreshToken string, accessToken string) error {
	definition := r.providers.StorageByKey(provider)
	if definition == nil {
		return fmt.Errorf("mailbox provider has no token storage: %s", provider)
	}
	fields, ok := definition.TokenFields()
	if !ok {
		return fmt.Errorf("mailbox provider has no token storage: %s", provider)
	}
	table, err := mailboxprovider.SQLIdentifier(fields.Table)
	if err != nil {
		return err
	}
	emailColumn, err := mailboxprovider.SQLIdentifier(fields.EmailColumn)
	if err != nil {
		return err
	}
	refreshColumn, err := mailboxprovider.SQLIdentifier(fields.RefreshTokenColumn)
	if err != nil {
		return err
	}
	accessColumn, err := mailboxprovider.SQLIdentifier(fields.AccessTokenColumn)
	if err != nil {
		return err
	}

	args := []any{strings.TrimSpace(refreshToken), strings.TrimSpace(accessToken)}
	assignments := []string{
		fmt.Sprintf("%s = $1", refreshColumn),
		fmt.Sprintf("%s = $2", accessColumn),
	}
	if fields.AuthStatusColumn != "" {
		column, err := mailboxprovider.SQLIdentifier(fields.AuthStatusColumn)
		if err != nil {
			return err
		}
		args = append(args, mailboxmodel.AuthStatusAuthorized)
		assignments = append(assignments, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	if fields.LastErrorColumn != "" {
		column, err := mailboxprovider.SQLIdentifier(fields.LastErrorColumn)
		if err != nil {
			return err
		}
		args = append(args, "")
		assignments = append(assignments, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	if fields.UpdatedAtColumn != "" {
		column, err := mailboxprovider.SQLIdentifier(fields.UpdatedAtColumn)
		if err != nil {
			return err
		}
		args = append(args, time.Now().Unix())
		assignments = append(assignments, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	args = append(args, emailx.Normalize(email))
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d", table, strings.Join(assignments, ", "), emailColumn, len(args))
	_, err = r.pool.Exec(ctx, query, args...)
	return err
}
