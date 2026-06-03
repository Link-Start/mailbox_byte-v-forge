package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/jackc/pgx/v5"
)

func (s *MailboxStore) DeleteMailbox(ctx context.Context, email string) (bool, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return false, errors.New("email_address is required")
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row, err := scanMailbox(tx.QueryRow(ctx, s.mailboxSelectSQL()+" WHERE m.email = $1 FOR UPDATE", email))
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

	deleteEmails := []string{row.Email}
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

func deleteMailboxInbox(ctx context.Context, tx pgx.Tx, emails []string) (bool, error) {
	args, inClause := sqlInArgs(emails)
	messageTag, err := tx.Exec(ctx, "DELETE FROM mailbox_inbox_messages WHERE mailbox_email IN ("+inClause+")", args...)
	if err != nil {
		return false, err
	}
	seenTag, err := tx.Exec(ctx, "DELETE FROM mailbox_inbox_seen WHERE mailbox_email IN ("+inClause+")", args...)
	if err != nil {
		return false, err
	}
	return messageTag.RowsAffected() > 0 || seenTag.RowsAffected() > 0, nil
}

func sqlInArgs(values []string) ([]any, string) {
	args := make([]any, 0, len(values))
	placeholders := make([]string, 0, len(values))
	for _, item := range values {
		args = append(args, item)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
	}
	return args, strings.Join(placeholders, ",")
}
