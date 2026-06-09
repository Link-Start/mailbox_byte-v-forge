package inboxapp

import (
	"context"
	"fmt"
	"time"

	commonv1 "mailboxapi/internal/contracts/commonv1"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/secretref"
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

func OTPSecretRef(message *mailboxv1.EmailInboxMessage, expiresAt time.Time) *commonv1.SecretRef {
	secretID := secretref.StableID(
		"mailbox-email-otp",
		message.GetProviderKey(),
		message.GetMailboxEmail(),
		message.GetId(),
		fmt.Sprintf("%d", message.GetReceivedAtUnix()),
	)
	return secretref.New("mailbox", "email_otp", secretID, expiresAt)
}

func ApplyOTPSecretRef(message *mailboxv1.EmailInboxMessage, ref *commonv1.SecretRef) {
	if message == nil || ref == nil {
		return
	}
	for _, signal := range message.GetSignals() {
		if signal.GetKind() == mailboxv1.EmailSignalKind_EMAIL_SIGNAL_KIND_OTP {
			signal.SecretRef = ref
		}
	}
	if message.GetPrimarySignal().GetKind() == mailboxv1.EmailSignalKind_EMAIL_SIGNAL_KIND_OTP {
		message.PrimarySignal.SecretRef = ref
	}
}
