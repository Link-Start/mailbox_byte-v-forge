package inboxapp

import (
	"strings"

	commonv1 "mailboxapi/internal/contracts/commonv1"
	"mailboxapi/internal/emailx"
)

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
