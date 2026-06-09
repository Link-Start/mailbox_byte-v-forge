package inboxapp

import (
	"context"
	"errors"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

func (s *Service) ListMessages(ctx context.Context, email string, limit int32) ([]*mailboxv1.EmailInboxMessage, error) {
	return s.ListMessagesSince(ctx, email, limit, 0)
}

func (s *Service) ListMessagesSince(ctx context.Context, email string, limit int32, receivedAfterUnix int64) ([]*mailboxv1.EmailInboxMessage, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	n := MessageLimitValue(limit, 25)
	rows, err := s.repo.ListInboxRows(ctx, email, n, receivedAfterUnix)
	if err != nil {
		return nil, err
	}

	out := []*mailboxv1.EmailInboxMessage{}
	for _, row := range rows {
		message := MessageFromRowLenient(row, s.normalizeProvider)
		if err := s.AttachSignalSecrets(ctx, message); err != nil {
			s.log("list mailbox email signal secret refresh failed email=%s: %v", emailx.Redact(email), err)
		}
		out = append(out, message)
	}
	return out, nil
}
