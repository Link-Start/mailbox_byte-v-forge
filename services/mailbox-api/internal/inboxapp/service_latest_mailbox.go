package inboxapp

import (
	"context"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

func (s *Service) latestMessageForMailbox(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, parserProfile string, signalKind mailboxv1.EmailSignalKind) (*mailboxv1.EmailInboxMessage, bool, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return nil, false, nil
	}
	rows, err := s.repo.LatestInboxRows(ctx, email, subjectKeyword, issuedAfterUnix, 50)
	if err != nil {
		return nil, false, err
	}

	for _, row := range rows {
		msg, err := MessageFromRow(row, parserProfile, s.normalizeProvider)
		if err != nil {
			return nil, false, err
		}
		if err := s.AttachSignalSecrets(ctx, msg); err != nil {
			s.log("latest mailbox email signal secret refresh failed email=%s: %v", emailx.Redact(email), err)
		}
		if MessageHasSignal(msg, signalKind) {
			return msg, true, nil
		}
	}
	return nil, false, nil
}
