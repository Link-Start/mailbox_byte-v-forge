package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
)

func (s *MailboxStore) ListInboxMessages(ctx context.Context, email string, limit int32) ([]*mailboxv1.EmailInboxMessage, error) {
	return s.ListInboxMessagesSince(ctx, email, limit, 0)
}

func (s *MailboxStore) ListInboxMessagesSince(ctx context.Context, email string, limit int32, receivedAfterUnix int64) ([]*mailboxv1.EmailInboxMessage, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	n := messageLimitValue(limit, defaultMessageLimit)
	args := []any{email}
	query := inboxMessageSelectSQL + "WHERE mailbox_email = $1"
	if receivedAfterUnix > 0 {
		args = append(args, receivedAfterUnix)
		query += fmt.Sprintf(" AND received_at > $%d", len(args))
	}
	args = append(args, n)
	query += fmt.Sprintf(`
		ORDER BY received_at DESC, updated_at DESC, message_key DESC
		LIMIT $%d
	`, len(args))
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*mailboxv1.EmailInboxMessage{}
	for rows.Next() {
		row, err := scanInboxMessageRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, inboxMessageToProtoLenient(row))
	}
	return out, rows.Err()
}
