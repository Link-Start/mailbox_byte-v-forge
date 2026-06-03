package main

import (
	"context"
	"fmt"
	"time"

	commonv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/common/v1"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/byte-v-forge/common-lib/secretref"
)

func (s *MailboxStore) attachEmailSignalSecrets(ctx context.Context, message *mailboxv1.EmailInboxMessage) error {
	if message == nil || s == nil || s.secrets == nil {
		return nil
	}
	code, _ := extractEmailOTP(message)
	if code == "" {
		return nil
	}
	expiresAt := s.emailOTPExpiresAt(message)
	ref := emailOTPSecretRef(message, expiresAt)
	applyEmailOTPSecretRef(message, ref)
	if ref == nil || !expiresAt.After(time.Now()) {
		return nil
	}
	return s.secrets.SaveOTP(ctx, ref, code, expiresAt)
}

func (s *MailboxStore) emailOTPExpiresAt(message *mailboxv1.EmailInboxMessage) time.Time {
	ttl := time.Hour
	if s != nil && s.secrets != nil && s.secrets.ttl > 0 {
		ttl = s.secrets.ttl
	}
	receivedAt := time.Unix(message.GetReceivedAtUnix(), 0)
	if message.GetReceivedAtUnix() <= 0 {
		receivedAt = time.Now()
	}
	return receivedAt.Add(ttl)
}

func emailOTPSecretRef(message *mailboxv1.EmailInboxMessage, expiresAt time.Time) *commonv1.SecretRef {
	secretID := secretref.StableID(
		"mailbox-email-otp",
		message.GetProviderKey(),
		message.GetMailboxEmail(),
		message.GetId(),
		fmt.Sprintf("%d", message.GetReceivedAtUnix()),
	)
	return secretref.New("mailbox", "email_otp", secretID, expiresAt)
}

func applyEmailOTPSecretRef(message *mailboxv1.EmailInboxMessage, ref *commonv1.SecretRef) {
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
