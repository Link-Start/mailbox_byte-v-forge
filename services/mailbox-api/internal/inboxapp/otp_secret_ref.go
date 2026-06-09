package inboxapp

import (
	"fmt"
	"time"

	commonv1 "mailboxapi/internal/contracts/commonv1"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/secretref"
)

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
