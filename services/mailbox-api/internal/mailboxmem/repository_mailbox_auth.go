package mailboxmem

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mailboxapi/internal/emailx"

	"mailboxapi/internal/mailboxmodel"
)

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
