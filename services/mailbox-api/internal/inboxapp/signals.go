package inboxapp

import mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

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
	if signal := message.GetPrimarySignal(); signal.GetKind() == kind {
		return true
	}
	for _, signal := range message.GetSignals() {
		if signal.GetKind() == kind {
			return true
		}
	}
	return false
}
