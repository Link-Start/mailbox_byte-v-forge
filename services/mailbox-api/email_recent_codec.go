package main

import (
	"strings"

	"google.golang.org/protobuf/proto"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/protojsonx"
)

func encodeRecentEmailMessage(message *mailboxv1.EmailInboxMessage) (string, bool) {
	if message == nil || emailx.Normalize(message.GetMailboxEmail()) == "" {
		return "", false
	}
	cloned, ok := proto.Clone(message).(*mailboxv1.EmailInboxMessage)
	if !ok {
		return "", false
	}
	cloned.MailboxEmail = emailx.Normalize(cloned.GetMailboxEmail())
	payload, err := protojsonx.Marshal(cloned)
	if err != nil {
		return "", false
	}
	return string(payload), true
}

func decodeRecentEmailMessage(payload string, parserProfile string) (*mailboxv1.EmailInboxMessage, bool) {
	message := &mailboxv1.EmailInboxMessage{}
	if err := protojsonx.Unmarshal([]byte(payload), message); err != nil {
		return nil, false
	}
	return inboxapp.MessageWithSignals(message, parserProfile), true
}

func recentEmailMatches(message *mailboxv1.EmailInboxMessage, subjectKeyword string, issuedAfterUnix int64, signalKind mailboxv1.EmailSignalKind) bool {
	if message == nil {
		return false
	}
	if issuedAfterUnix > 0 && message.GetReceivedAtUnix() < issuedAfterUnix {
		return false
	}
	if keyword := strings.ToLower(strings.TrimSpace(subjectKeyword)); keyword != "" && !recentEmailContainsKeyword(message, keyword) {
		return false
	}
	return inboxapp.MessageHasSignal(message, signalKind)
}

func recentEmailContainsKeyword(message *mailboxv1.EmailInboxMessage, keyword string) bool {
	return strings.Contains(strings.ToLower(message.GetSubject()), keyword) ||
		strings.Contains(strings.ToLower(message.GetBodyPreview()), keyword)
}
