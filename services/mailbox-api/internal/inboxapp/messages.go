package inboxapp

import (
	"encoding/json"
	"html"
	"regexp"
	"strings"
	"time"

	commonv1 "mailboxapi/internal/contracts/commonv1"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/hashx"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func MessageFromRow(row MessageRow, profile string, normalizeProvider func(string) string) (*mailboxv1.EmailInboxMessage, error) {
	recipients := []string{}
	if strings.TrimSpace(row.RecipientsJSON) != "" {
		if err := json.Unmarshal([]byte(row.RecipientsJSON), &recipients); err != nil {
			return nil, err
		}
	}
	return messageFromRow(row, profile, recipients, normalizeProvider), nil
}

func MessageFromRowLenient(row MessageRow, normalizeProvider func(string) string) *mailboxv1.EmailInboxMessage {
	recipients := []string{}
	if err := json.Unmarshal([]byte(row.RecipientsJSON), &recipients); err != nil {
		recipients = []string{}
	}
	return messageFromRow(row, "", recipients, normalizeProvider)
}

func messageFromRow(row MessageRow, profile string, recipients []string, normalizeProvider func(string) string) *mailboxv1.EmailInboxMessage {
	if normalizeProvider == nil {
		normalizeProvider = NormalizeProviderKey
	}
	return MessageWithSignals(&mailboxv1.EmailInboxMessage{
		Id:                 row.ID,
		MailboxEmail:       emailx.Normalize(row.MailboxEmail),
		Subject:            row.Subject,
		FromAddress:        row.FromAddress,
		BodyPreview:        row.BodyPreview,
		ReceivedAtUnix:     row.ReceivedAtUnix,
		Recipients:         UniqueEmails(recipients),
		ProviderKey:        normalizeProvider(row.Provider),
		SourceMailboxEmail: emailx.Normalize(row.SourceEmail),
		BodyArtifactRef:    ArtifactRef(row.Provider, row.MailboxEmail, row.ID, "body_text", int64(len(row.BodyText)), normalizeProvider),
		HtmlArtifactRef:    ArtifactRef(row.Provider, row.MailboxEmail, row.ID, "html_body", int64(len(row.HTMLBody)), normalizeProvider),
		RawSize:            row.RawSize,
	}, profile)
}

func ArtifactRef(provider string, mailboxEmail string, messageID string, purpose string, sizeBytes int64, normalizeProvider func(string) string) *commonv1.ArtifactRef {
	if normalizeProvider == nil {
		normalizeProvider = NormalizeProviderKey
	}
	provider = normalizeProvider(provider)
	mailboxEmail = emailx.Normalize(mailboxEmail)
	messageID = strings.TrimSpace(messageID)
	purpose = strings.TrimSpace(purpose)
	if provider == "" || mailboxEmail == "" || messageID == "" || purpose == "" || sizeBytes <= 0 {
		return nil
	}
	artifactID := strings.Join([]string{"mailbox", provider, mailboxEmail, messageID, purpose}, ":")
	return &commonv1.ArtifactRef{
		ArtifactId:  artifactID,
		Uri:         "mailbox://inbox/" + artifactID,
		ContentType: artifactContentType(purpose),
		SizeBytes:   sizeBytes,
		Purpose:     purpose,
	}
}

func artifactContentType(purpose string) string {
	if purpose == "html_body" {
		return "text/html"
	}
	return "text/plain"
}

func UniqueEmails(values []string) []string {
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

func CompactMessageText(value string, limit int) string {
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

func MessageLimitValue(limit int32, fallback int) int {
	n := int(limit)
	if n <= 0 {
		n = fallback
	}
	if n <= 0 {
		n = 25
	}
	if n > 100 {
		n = 100
	}
	return n
}

func InboxReceivedAfter(watermarkNs int64, overlapSeconds int) int64 {
	if watermarkNs <= 0 {
		return 0
	}
	after := watermarkNs - int64(overlapSeconds)*int64(time.Second)
	if after < 0 {
		return 0
	}
	return after
}

func MessageMailboxEmails(accountEmail string, recipients []string) []string {
	items := []string{emailx.Normalize(accountEmail)}
	for _, recipient := range recipients {
		if email := emailx.Normalize(recipient); email != "" {
			items = append(items, email)
		}
	}
	return UniqueEmails(items)
}

func StableMessageKey(provider string, mailboxEmail string, value string) string {
	return hashx.StableParts(NormalizeProviderKey(provider), emailx.Normalize(mailboxEmail), strings.TrimSpace(value))
}

func NormalizeProviderKey(provider string) string {
	return strings.Trim(strings.ToLower(strings.TrimSpace(provider)), ".")
}
