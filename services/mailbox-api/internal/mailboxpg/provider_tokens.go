package mailboxpg

import (
	"context"
	"fmt"

	"mailboxapi/internal/emailx"
)

func (r *Repository) updateMailboxTokens(ctx context.Context, provider string, email string, refreshToken string, accessToken string) error {
	statement, err := r.updateMailboxTokenStatement(provider, refreshToken, accessToken)
	if err != nil {
		return err
	}
	statement.args = append(statement.args, emailx.Normalize(email))
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d", statement.table, statement.assignmentsSQL(), statement.emailColumn, len(statement.args))
	_, err = r.pool.Exec(ctx, query, statement.args...)
	return err
}
