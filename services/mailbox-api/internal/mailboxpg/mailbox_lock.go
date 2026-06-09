package mailboxpg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"
)

func lockStoredMailbox(ctx context.Context, tx pgx.Tx, email string) (string, error) {
	var provider string
	err := tx.QueryRow(ctx, "SELECT provider FROM mailboxes WHERE email = $1 FOR UPDATE", emailx.Normalize(email)).Scan(&provider)
	if err != nil {
		return "", err
	}
	return provider, nil
}
