package mailboxpg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxprovider"
)

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
