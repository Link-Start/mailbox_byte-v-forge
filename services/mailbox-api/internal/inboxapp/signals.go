package inboxapp

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	commonv1 "mailboxapi/internal/contracts/commonv1"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/secretref"
)

var (
	emailOTPContextPattern    = regexp.MustCompile(`(?i)(?:verification|security|login|one[- ]?time|otp|code|验证码|安全代码)[^0-9]{0,80}([0-9]{4,8})`)
	emailOTPStandalonePattern = regexp.MustCompile(`(^|[^0-9])([0-9]{6})([^0-9]|$)`)
)

func MessageWithSignals(message *mailboxv1.EmailInboxMessage, _ string) *mailboxv1.EmailInboxMessage {
	if message == nil {
		return nil
	}
	code, evidence := ExtractEmailOTP(message)
	if code == "" {
		message.Signals = nil
		message.PrimarySignal = nil
		return message
	}
	signal := &mailboxv1.EmailSignal{
		Kind:            mailboxv1.EmailSignalKind_EMAIL_SIGNAL_KIND_OTP,
		SecretRef:       OTPSecretRef(message, time.Time{}),
		Label:           "verification_code",
		Profile:         "generic",
		Parser:          "mailbox-email-otp",
		Confidence:      70,
		EvidencePreview: evidence,
	}
	message.Signals = []*mailboxv1.EmailSignal{signal}
	message.PrimarySignal = signal
	return message
}

func MessageHasSignal(message *mailboxv1.EmailInboxMessage, kind mailboxv1.EmailSignalKind) bool {
	if message == nil {
		return false
	}
	if kind == mailboxv1.EmailSignalKind_EMAIL_SIGNAL_KIND_UNSPECIFIED {
		return true
	}
	if signal := message.GetPrimarySignal(); signal.GetKind() == kind && signal.GetSecretRef().GetSecretId() != "" {
		return true
	}
	for _, signal := range message.GetSignals() {
		if signal.GetKind() == kind && signal.GetSecretRef().GetSecretId() != "" {
			return true
		}
	}
	return false
}

func ExtractEmailOTP(message *mailboxv1.EmailInboxMessage) (string, string) {
	if message == nil {
		return "", ""
	}
	text := strings.Join([]string{
		message.GetSubject(),
		message.GetFromAddress(),
		message.GetBodyPreview(),
	}, "\n")
	if match := emailOTPContextPattern.FindStringSubmatch(text); len(match) >= 2 {
		return NormalizeEmailOTP(match[1]), strings.TrimSpace(match[0])
	}
	if match := emailOTPStandalonePattern.FindStringSubmatch(text); len(match) >= 3 {
		return NormalizeEmailOTP(match[2]), strings.TrimSpace(match[0])
	}
	return "", ""
}

func NormalizeEmailOTP(value string) string {
	return strings.TrimSpace(strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r", "", "-", "").Replace(value))
}

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
	ApplyOTPSecretRef(message, ref)
	if ref == nil || !expiresAt.After(s.now()) {
		return nil
	}
	return s.secrets.SaveOTP(ctx, ref, code, expiresAt)
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
