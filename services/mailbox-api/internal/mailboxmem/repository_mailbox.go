package mailboxmem

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
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
