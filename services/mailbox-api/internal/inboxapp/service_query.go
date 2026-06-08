package inboxapp

import (
	"context"
	"errors"
	"time"

	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
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

func (s *Service) latestCachedMessage(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, parserProfile string, signalKind mailboxv1.EmailSignalKind) (*mailboxv1.EmailInboxMessage, bool, error) {
	if s == nil || s.recent == nil {
		return nil, false, nil
	}
	return s.recent.Latest(ctx, email, subjectKeyword, issuedAfterUnix, parserProfile, signalKind)
}

func (s *Service) recordRecentInboxMessages(ctx context.Context, messages []*mailboxv1.EmailInboxMessage) {
	if s == nil || s.recent == nil || len(messages) == 0 {
		return
	}
	cacheCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	if err := s.recent.Record(cacheCtx, messages); err != nil {
		s.log("record recent mailbox email cache failed: %v", err)
	}
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
