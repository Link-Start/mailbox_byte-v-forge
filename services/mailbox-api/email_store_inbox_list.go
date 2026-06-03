package main

import (
	"context"
	"errors"

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
	rows, err := s.mailboxes.ListInboxRows(ctx, email, n, receivedAfterUnix)
	if err != nil {
		return nil, err
	}

	out := []*mailboxv1.EmailInboxMessage{}
	for _, row := range rows {
		message := inboxMessageToProtoLenient(row)
		if err := s.attachEmailSignalSecrets(ctx, message); err != nil {
			logWarning("list mailbox email signal secret refresh failed email=%s: %v", emailx.Redact(email), err)
		}
		out = append(out, message)
	}
	return out, nil
}
