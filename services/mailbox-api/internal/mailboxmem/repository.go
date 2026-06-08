package mailboxmem

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/pagex"
	"mailboxapi/internal/redactx"
	"mailboxapi/internal/stringx"

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

func (r *Repository) UpsertMailbox(ctx context.Context, mailbox *mailboxmodel.Record) (*mailboxmodel.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if mailbox == nil {
		return nil, errors.New("mailbox is required")
	}
	email := emailx.Normalize(mailbox.GetEmailAddress())
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	requestedProvider := r.providers.NormalizeProviderInput(mailbox.GetProviderKey())
	now := time.Now().Unix()

	r.mu.Lock()
	entry, exists := r.mailboxes[email]
	record := cloneRecord(entry.record)
	if record == nil {
		record = &mailboxmodel.Record{EmailAddress: email, CreatedAt: now}
	}
	if requestedProvider != "" || record.ProviderKey == "" {
		record.ProviderKey = stringx.FirstNonEmpty(requestedProvider, r.providers.DefaultKey())
	}
	mergeCredentials(record, mailbox, exists)
	record.EmailAddress = email
	record.Domain = domainForEmail(email)
	record.UpdatedAt = now
	entry.record = record
	r.mailboxes[email] = entry
	r.mu.Unlock()

	return r.FindMailbox(ctx, email)
}

func (r *Repository) MarkEmailAuthStatus(ctx context.Context, email string, authStatus string, lastError string) (*mailboxmodel.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	email = emailx.Normalize(email)
	authStatus = strings.TrimSpace(authStatus)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	if authStatus == "" {
		return nil, errors.New("auth_status is required")
	}
	r.mu.Lock()
	entry, ok := r.mailboxes[email]
	if !ok || entry.record == nil {
		r.mu.Unlock()
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	entry.record.AuthStatus = authStatus
	entry.record.LastError = safeText(lastError)
	entry.record.UpdatedAt = time.Now().Unix()
	r.mailboxes[email] = entry
	r.mu.Unlock()
	return r.FindMailbox(ctx, email)
}

func (r *Repository) UpdateMailboxTokens(ctx context.Context, email string, refreshToken string, accessToken string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	email = emailx.Normalize(email)
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.mailboxes[email]
	if !ok || entry.record == nil {
		return fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	definition := r.providers.StorageByKey(entry.record.ProviderKey)
	if definition == nil {
		return fmt.Errorf("mailbox provider has no token storage: %s", entry.record.ProviderKey)
	}
	if _, ok := definition.TokenFields(); !ok {
		return fmt.Errorf("mailbox provider has no token storage: %s", entry.record.ProviderKey)
	}
	entry.record.RefreshToken = strings.TrimSpace(refreshToken)
	entry.record.AccessToken = strings.TrimSpace(accessToken)
	entry.record.AuthStatus = mailboxmodel.AuthStatusAuthorized
	entry.record.LastError = ""
	entry.record.UpdatedAt = time.Now().Unix()
	r.mailboxes[email] = entry
	return nil
}

func (r *Repository) DeleteMailbox(ctx context.Context, email string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	email = emailx.Normalize(email)
	if email == "" {
		return false, errors.New("email_address is required")
	}
	r.mu.Lock()
	_, mailboxExists := r.mailboxes[email]
	delete(r.mailboxes, email)
	messageDeleted := r.deleteInboxLocked(email)
	r.mu.Unlock()
	return mailboxExists || messageDeleted, nil
}

func (r *Repository) FindMailbox(ctx context.Context, email string) (*mailboxmodel.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	email = emailx.Normalize(email)
	r.mu.RLock()
	record := cloneRecord(r.mailboxes[email].record)
	r.mu.RUnlock()
	if record == nil {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	return r.project(record), nil
}

func (r *Repository) PollMailboxForEmail(ctx context.Context, email string) (*mailboxmodel.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	email = emailx.Normalize(email)
	r.mu.RLock()
	record := cloneRecord(r.mailboxes[email].record)
	if record == nil {
		canonical := emailx.CanonicalPlusAlias(email)
		if canonical != "" && canonical != email {
			record = cloneRecord(r.mailboxes[canonical].record)
		}
	}
	r.mu.RUnlock()
	if record == nil {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	if err := r.providers.ValidatePoll(providerRecord(record)); err != nil {
		return nil, err
	}
	return r.project(record), nil
}

func (r *Repository) ListMailboxes(ctx context.Context, authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxmodel.ListPage, error) {
	if err := ctx.Err(); err != nil {
		return mailboxmodel.ListPage{}, err
	}
	cursor, err := pagex.DecodeKeysetCursor(cursorValue)
	if err != nil {
		return mailboxmodel.ListPage{}, mailboxmodel.ErrInvalidMailboxListCursor
	}
	query := mailboxprovider.ListQuery{
		AuthStatus:   strings.TrimSpace(authStatus),
		Provider:     r.providers.NormalizeProviderInput(provider),
		EmailAddress: emailx.Normalize(emailAddress),
		Cursor:       cursor,
		Limit:        pagex.NormalizePageLimit(int(limit)),
	}
	r.mu.RLock()
	rows := r.listStoredMailboxesLocked(query)
	rows = append(rows, r.listVirtualMailboxesLocked(query)...)
	r.mu.RUnlock()
	return mailboxPageFromRows(rows, query.Limit), nil
}

func (r *Repository) ListOAuthMailboxes(ctx context.Context, limit int32) ([]*mailboxmodel.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n := int(limit)
	if n <= 0 {
		n = 100
	}
	if n > 500 {
		n = 500
	}
	r.mu.RLock()
	rows := make([]*mailboxmodel.Record, 0, len(r.mailboxes))
	for _, entry := range r.mailboxes {
		record := cloneRecord(entry.record)
		if record == nil || record.AuthStatus != mailboxmodel.AuthStatusAuthorized {
			continue
		}
		if r.providers.ValidatePoll(providerRecord(record)) != nil {
			continue
		}
		rows = append(rows, r.project(record))
	}
	r.mu.RUnlock()
	sortMailboxRows(rows)
	if len(rows) > n {
		rows = rows[:n]
	}
	return rows, nil
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
