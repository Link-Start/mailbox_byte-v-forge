package mailboxmem

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
)

func (r *Repository) InboxWatermark(ctx context.Context, email string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	email = emailx.Normalize(email)
	if email == "" {
		return 0, errors.New("email_address is required")
	}
	r.mu.RLock()
	entry, ok := r.mailboxes[email]
	r.mu.RUnlock()
	if !ok || entry.record == nil {
		return 0, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	return entry.inboxWatermark, nil
}

func (r *Repository) HasInboxMessages(ctx context.Context, email string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	email = emailx.Normalize(email)
	if email == "" {
		return false, errors.New("email_address is required")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, message := range r.messages {
		if message.row.MailboxEmail == email {
			return true, nil
		}
	}
	return false, nil
}

func (r *Repository) ListInboxRows(ctx context.Context, email string, limit int, receivedAfterUnix int64) ([]inboxapp.MessageRow, error) {
	return r.filterInboxRows(ctx, email, "", receivedAfterUnix, false, limit)
}

func (r *Repository) GetInboxRow(ctx context.Context, email string, messageID string, provider string) (inboxapp.MessageRow, bool, error) {
	if err := ctx.Err(); err != nil {
		return inboxapp.MessageRow{}, false, err
	}
	email = emailx.Normalize(email)
	if email == "" {
		return inboxapp.MessageRow{}, false, errors.New("email_address is required")
	}
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return inboxapp.MessageRow{}, false, errors.New("message_id is required")
	}
	provider = r.providers.NormalizeProviderInput(provider)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, message := range r.messages {
		if message.row.MailboxEmail != email || message.row.ID != messageID {
			continue
		}
		if provider != "" && message.row.Provider != provider {
			continue
		}
		return message.row, true, nil
	}
	return inboxapp.MessageRow{}, false, nil
}

func (r *Repository) LatestInboxRows(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, limit int) ([]inboxapp.MessageRow, error) {
	return r.filterInboxRows(ctx, email, subjectKeyword, issuedAfterUnix, true, limit)
}

func (r *Repository) filterInboxRows(ctx context.Context, email string, keyword string, timestamp int64, includeEqual bool, limit int) ([]inboxapp.MessageRow, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	email = emailx.Normalize(email)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	r.mu.RLock()
	rows := []storedMessage{}
	for _, message := range r.messages {
		if message.row.MailboxEmail != email {
			continue
		}
		if includeEqual {
			if message.row.ReceivedAtUnix < timestamp {
				continue
			}
			if !messageMatchesKeyword(message.row, keyword) {
				continue
			}
		} else if message.row.ReceivedAtUnix <= timestamp {
			continue
		}
		rows = append(rows, message)
	}
	r.mu.RUnlock()
	sortStoredMessages(rows)
	limit = normalizeInboxRowLimit(limit)
	if len(rows) > limit {
		rows = rows[:limit]
	}
	out := make([]inboxapp.MessageRow, 0, len(rows))
	for _, message := range rows {
		out = append(out, message.row)
	}
	return out, nil
}

func messageMatchesKeyword(row inboxapp.MessageRow, keyword string) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	return strings.Contains(strings.ToLower(row.Subject), keyword) ||
		strings.Contains(strings.ToLower(row.BodyPreview), keyword) ||
		strings.Contains(strings.ToLower(row.BodyText), keyword)
}

func sortStoredMessages(rows []storedMessage) {
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].row.ReceivedAtUnix == rows[j].row.ReceivedAtUnix {
			if rows[i].updatedAt == rows[j].updatedAt {
				return rows[i].key > rows[j].key
			}
			return rows[i].updatedAt > rows[j].updatedAt
		}
		return rows[i].row.ReceivedAtUnix > rows[j].row.ReceivedAtUnix
	})
}

func normalizeInboxRowLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	return limit
}
