package inboxapp

import (
	"context"
	"time"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

func (s *Service) prepareUnseenMessage(ctx context.Context, message *mailboxv1.EmailInboxMessage) error {
	MessageWithSignals(message, "")
	return s.AttachSignalSecrets(ctx, message)
}

func (s *Service) AttachSignalSecrets(ctx context.Context, message *mailboxv1.EmailInboxMessage) error {
	if message == nil || s == nil || s.secrets == nil {
		return nil
	}
	code, _ := ExtractEmailOTP(message)
	if code == "" {
		return nil
	}
	expiresAt := s.emailOTPExpiresAt(message)
	ref := OTPSecretRef(message, expiresAt)
	if ref == nil || !expiresAt.After(s.now()) {
		return nil
	}
	if err := s.secrets.SaveOTP(ctx, ref, code, expiresAt); err != nil {
		return err
	}
	ApplyOTPSecretRef(message, ref)
	return nil
}

func (s *Service) emailOTPExpiresAt(message *mailboxv1.EmailInboxMessage) time.Time {
	ttl := time.Hour
	if s != nil && s.secretTTL > 0 {
		ttl = s.secretTTL
	}
	receivedAt := time.Unix(message.GetReceivedAtUnix(), 0)
	if message.GetReceivedAtUnix() <= 0 {
		receivedAt = s.now()
	}
	return receivedAt.Add(ttl)
}
