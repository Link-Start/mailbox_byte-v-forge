package mailboxmem

import (
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/redactx"
)

const mailboxErrorSnippetLimit = 600

func (r *Repository) project(record *mailboxmodel.Record) *mailboxmodel.Record {
	record = cloneRecord(record)
	if record == nil {
		return nil
	}
	record.ProviderKey = r.providers.NormalizeProviderInput(record.GetProviderKey())
	record.Domain = domainForEmail(record.GetEmailAddress())
	r.providers.PrepareProjection(record)
	return record
}

func domainForEmail(email string) string {
	_, domain, ok := strings.Cut(emailx.Normalize(email), "@")
	if !ok {
		return ""
	}
	return domain
}

func safeText(value string) string {
	return redactx.TextSnippet(value, mailboxErrorSnippetLimit)
}
