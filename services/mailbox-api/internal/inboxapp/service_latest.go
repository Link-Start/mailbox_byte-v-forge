package inboxapp

import (
	"context"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

func (s *Service) LatestMessage(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64) (*mailboxv1.EmailInboxMessage, bool, error) {
	return s.LatestMessageWithSignal(ctx, email, subjectKeyword, issuedAfterUnix, "", mailboxv1.EmailSignalKind_EMAIL_SIGNAL_KIND_UNSPECIFIED)
}

func (s *Service) LatestMessageWithSignal(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, parserProfile string, signalKind mailboxv1.EmailSignalKind) (*mailboxv1.EmailInboxMessage, bool, error) {
	candidates := UniqueEmails([]string{email, emailx.CanonicalPlusAlias(email)})
	for _, candidate := range candidates {
		msg, ok, err := s.latestCachedMessage(ctx, candidate, subjectKeyword, issuedAfterUnix, parserProfile, signalKind)
		if err != nil {
			s.log("latest mailbox email cache read failed email=%s: %v", emailx.Redact(candidate), err)
			break
		}
		if ok {
			if err := s.AttachSignalSecrets(ctx, msg); err != nil {
				s.log("latest mailbox email signal secret refresh failed email=%s: %v", emailx.Redact(candidate), err)
			}
			return msg, true, nil
		}
	}
	for _, candidate := range candidates {
		msg, ok, err := s.latestMessageForMailbox(ctx, candidate, subjectKeyword, issuedAfterUnix, parserProfile, signalKind)
		if err != nil || ok {
			return msg, ok, err
		}
	}
	return nil, false, nil
}

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
