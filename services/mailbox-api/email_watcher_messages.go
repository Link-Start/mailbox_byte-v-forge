package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/hashx"
	"mailboxapi/internal/timex"

	"mailboxapi/internal/inboxapp"
)

func inboxMessages(providerKey string, mailboxEmail string, messages []graphMessage) []*mailboxv1.EmailInboxMessage {
	out := make([]*mailboxv1.EmailInboxMessage, 0, len(messages))
	for _, msg := range messages {
		out = append(out, inboxMessage(providerKey, mailboxEmail, msg))
	}
	return out
}

func inboxMessage(providerKey string, mailboxEmail string, msg graphMessage) *mailboxv1.EmailInboxMessage {
	bodyPreview := strings.TrimSpace(msg.BodyPreview)
	if bodyPreview == "" {
		bodyPreview = inboxapp.CompactMessageText(msg.Body.Content, 500)
	}
	return &mailboxv1.EmailInboxMessage{
		Id:                 msg.ID,
		MailboxEmail:       emailx.Normalize(mailboxEmail),
		Subject:            strings.TrimSpace(msg.Subject),
		FromAddress:        strings.TrimSpace(msg.From.EmailAddress.Address),
		BodyPreview:        inboxapp.CompactMessageText(bodyPreview, 500),
		ReceivedAtUnix:     int64(timex.UnixFloat(msg.ReceivedDateTime)),
		Recipients:         inboxapp.UniqueEmails(messageAddresses(msg)),
		ProviderKey:        providerKey,
		SourceMailboxEmail: emailx.Normalize(mailboxEmail),
	}
}

func messageKey(msg graphMessage) string {
	if msg.ID != "" {
		return msg.ID
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hashx.SHA256Hex(string(raw))
}

func messageAddresses(msg graphMessage) []string {
	out := []string{}
	for _, list := range [][]graphRecipient{msg.ToRecipients, msg.CcRecipients, msg.BccRecipients} {
		for _, recipient := range list {
			if address := strings.TrimSpace(recipient.EmailAddress.Address); address != "" {
				out = append(out, address)
			}
		}
	}
	for _, header := range msg.InternetMessageHeaders {
		name := strings.ToLower(strings.TrimSpace(header.Name))
		value := header.Value
		if recipientHeaders[name] {
			out = append(out, emailPattern.FindAllString(value, -1)...)
			continue
		}
		if name == "received" {
			idx := strings.LastIndex(strings.ToLower(value), " for ")
			if idx >= 0 {
				out = append(out, emailPattern.FindAllString(value[idx+5:], -1)...)
			}
		}
	}
	return out
}

var recipientHeaders = map[string]bool{
	"to":                   true,
	"cc":                   true,
	"bcc":                  true,
	"delivered-to":         true,
	"envelope-to":          true,
	"x-envelope-to":        true,
	"x-original-to":        true,
	"x-original-recipient": true,
	"resent-to":            true,
	"apparently-to":        true,
	"x-forwarded-to":       true,
	"x-ms-exchange-organization-originalrecipient":          true,
	"x-ms-exchange-organization-originalenveloperecipients": true,
}
