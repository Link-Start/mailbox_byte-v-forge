package mailboxpg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

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
