package main

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/byte-v-forge/common-lib/hashx"
	"github.com/byte-v-forge/common-lib/timex"
)

func inboxMessages(mailboxEmail string, messages []graphMessage) []*mailboxv1.EmailInboxMessage {
	out := make([]*mailboxv1.EmailInboxMessage, 0, len(messages))
	for _, msg := range messages {
		out = append(out, inboxMessage(mailboxEmail, msg))
	}
	return out
}

func inboxMessage(mailboxEmail string, msg graphMessage) *mailboxv1.EmailInboxMessage {
	bodyPreview := strings.TrimSpace(msg.BodyPreview)
	if bodyPreview == "" {
		bodyPreview = compactMessageText(msg.Body.Content, 500)
	}
	body := msg.BodyPreview + "\n" + msg.Body.Content
	return &mailboxv1.EmailInboxMessage{
		Id:                 msg.ID,
		MailboxEmail:       emailx.Normalize(mailboxEmail),
		Subject:            strings.TrimSpace(msg.Subject),
		FromAddress:        strings.TrimSpace(msg.From.EmailAddress.Address),
		BodyPreview:        compactMessageText(bodyPreview, 500),
		ReceivedAtUnix:     int64(timex.UnixFloat(msg.ReceivedDateTime)),
		Recipients:         uniqueStrings(messageAddresses(msg)),
		ProviderKey:        emailProviderOutlook,
		SourceMailboxEmail: emailx.Normalize(mailboxEmail),
		BodyText:           compactMessageText(body, 5000),
	}
}

func compactMessageText(value string, limit int) string {
	text := htmlTagPattern.ReplaceAllString(html.UnescapeString(value), " ")
	text = strings.Join(strings.Fields(strings.ReplaceAll(text, "\u00a0", " ")), " ")
	if limit > 0 && len(text) > limit {
		runes := []rune(text)
		if len(runes) > limit {
			return string(runes[:limit])
		}
	}
	return text
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		trimmed := emailx.Normalize(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func messageLimitValue(limit int32, fallback int) int {
	n := int(limit)
	if n <= 0 {
		n = fallback
	}
	if n <= 0 {
		n = defaultMessageLimit
	}
	if n > 100 {
		n = 100
	}
	return n
}

func inboxReceivedAfter(watermarkNs int64, overlapSeconds int) int64 {
	if watermarkNs <= 0 {
		return 0
	}
	after := watermarkNs - int64(overlapSeconds)*int64(time.Second)
	if after < 0 {
		return 0
	}
	return after
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
