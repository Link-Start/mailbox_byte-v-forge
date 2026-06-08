package mailboxmem

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
	"mailboxapi/internal/pagex"
	"mailboxapi/internal/stringx"
)

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
