package mailboxpg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"
)

func (r *Repository) DeleteMailbox(ctx context.Context, email string) (bool, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return false, errors.New("email_address is required")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = lockStoredMailbox(ctx, tx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		deleted, deleteErr := deleteMailboxInbox(ctx, tx, []string{email})
		if deleteErr != nil {
			return false, deleteErr
		}
		if err := tx.Commit(ctx); err != nil {
			return false, err
		}
		return deleted, nil
	}
	if err != nil {
		return false, err
	}

	deleteEmails := []string{email}
	if _, err := deleteMailboxInbox(ctx, tx, deleteEmails); err != nil {
		return false, err
	}
	args, inClause := sqlInArgs(deleteEmails)
	tag, err := tx.Exec(ctx, "DELETE FROM mailboxes WHERE email IN ("+inClause+")", args...)
	if err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
