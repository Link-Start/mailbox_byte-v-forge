package mailboxmem

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
)

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
