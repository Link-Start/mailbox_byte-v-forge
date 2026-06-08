package mailboxmem

import (
	"context"
	"errors"
	"strings"
	"sync"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/redactx"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

const mailboxErrorSnippetLimit = 600

type Repository struct {
	mu        sync.RWMutex
	providers *mailboxprovider.Registry
	mailboxes map[string]mailboxEntry
	messages  map[string]storedMessage
}

type mailboxEntry struct {
	record         *mailboxmodel.Record
	inboxWatermark int64
}

type storedMessage struct {
	key       string
	createdAt int64
	updatedAt int64
	row       inboxapp.MessageRow
}

func NewRepository(providers *mailboxprovider.Registry) (*Repository, error) {
	if providers == nil {
		return nil, errors.New("mailbox providers are required")
	}
	return &Repository{
		providers: providers,
		mailboxes: map[string]mailboxEntry{},
		messages:  map[string]storedMessage{},
	}, nil
}

func (r *Repository) Close() {}

func (r *Repository) RunOutboxWorker(context.Context, string, eventbus.Publisher, func(string, ...any)) error {
	return nil
}

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

func cloneRecord(record *mailboxmodel.Record) *mailboxmodel.Record {
	if record == nil {
		return nil
	}
	return &mailboxmodel.Record{
		EmailAddress: record.EmailAddress,
		Password:     record.Password,
		RefreshToken: record.RefreshToken,
		AccessToken:  record.AccessToken,
		LastError:    record.LastError,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
		AuthStatus:   record.AuthStatus,
		ProviderKey:  record.ProviderKey,
		LatestSignal: record.LatestSignal,
		Domain:       record.Domain,
	}
}

func providerRecord(record *mailboxmodel.Record) mailboxprovider.MailboxRecord {
	if record == nil {
		return mailboxprovider.MailboxRecord{}
	}
	return mailboxprovider.MailboxRecord{
		Email:        record.EmailAddress,
		Provider:     record.ProviderKey,
		RefreshToken: record.RefreshToken,
		AuthStatus:   record.AuthStatus,
	}
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
