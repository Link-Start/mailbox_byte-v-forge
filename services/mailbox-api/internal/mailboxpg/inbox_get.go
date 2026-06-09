package mailboxpg

import (
	"context"
	"errors"
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
)

func (r *Repository) GetInboxRow(ctx context.Context, email string, messageID string, provider string) (inboxapp.MessageRow, bool, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return inboxapp.MessageRow{}, false, errors.New("email_address is required")
	}
	if strings.TrimSpace(messageID) == "" {
		return inboxapp.MessageRow{}, false, errors.New("message_id is required")
	}
	rows, err := newInboxMessageQuery().
		WhereMailbox(email).
		WhereMessageID(messageID).
		WhereProvider(r.providers.NormalizeProviderInput(provider)).
		Limit(1).
		Query(ctx, r.pool)
	if err != nil {
		return inboxapp.MessageRow{}, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return inboxapp.MessageRow{}, false, rows.Err()
	}
	row, err := scanInboxMessageRow(rows)
	if err != nil {
		return inboxapp.MessageRow{}, false, err
	}
	return row, true, rows.Err()
}
