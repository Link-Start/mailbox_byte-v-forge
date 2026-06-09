package mailboxpg

import (
	"context"
	"time"

	"mailboxapi/internal/emailx"
)

func (r *Repository) UpdateMailboxTokens(ctx context.Context, email string, refreshToken string, accessToken string) error {
	email = emailx.Normalize(email)
	row, err := r.newMailboxSelectQuery().WhereEmail(email).ScanOne(ctx, r.pool)
	if err != nil {
		return err
	}
	if err := r.updateMailboxTokens(ctx, row.Provider, email, refreshToken, accessToken); err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, "UPDATE mailboxes SET updated_at = $1 WHERE email = $2", time.Now().Unix(), email)
	return err
}
