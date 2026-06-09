package mailboxmem

import (
	"context"
	"errors"

	"mailboxapi/internal/emailx"
)

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
