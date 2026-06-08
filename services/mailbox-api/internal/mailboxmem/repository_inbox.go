package mailboxmem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxprovider"
	"mailboxapi/internal/stringx"
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

func (r *Repository) RecordMessages(ctx context.Context, request inboxapp.RecordMessagesRequest) ([]*mailboxv1.EmailInboxMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	provider := r.providers.NormalizeProviderInput(request.Provider)
	if provider == "" {
		return nil, errors.New("email provider is required")
	}
	if len(request.Messages) == 0 {
		return []*mailboxv1.EmailInboxMessage{}, nil
	}
	now := time.Now().Unix()
	unseen := []*mailboxv1.EmailInboxMessage{}
	touchedMailboxes := map[string]struct{}{}
	touchedDomains := map[string]struct{}{}

	r.mu.Lock()
	for _, input := range request.Messages {
		if input.Message == nil {
			continue
		}
		for _, mailboxEmail := range persistTargetMailboxes(input, request.ExpandRecipients) {
			touchedMailboxes[mailboxEmail] = struct{}{}
			if domain := domainForEmail(mailboxEmail); domain != "" {
				touchedDomains[domain] = struct{}{}
			}
			persisted, key, row, err := r.prepareInboxMessage(provider, mailboxEmail, input, now)
			if err != nil {
				r.mu.Unlock()
				return nil, err
			}
			storageKey := messageStorageKey(provider, mailboxEmail, key)
			_, existed := r.messages[storageKey]
			r.messages[storageKey] = storedMessage{key: key, createdAt: now, updatedAt: now, row: row}
			r.trackInboxWatermarkLocked(mailboxEmail, persisted.GetReceivedAtUnix(), now)
			if !existed {
				if request.PrepareUnseen != nil {
					if err := request.PrepareUnseen(ctx, persisted); err != nil {
						r.mu.Unlock()
						return nil, err
					}
				}
				unseen = append(unseen, persisted)
			}
		}
	}
	r.pruneInboundLocked(provider, mailboxprovider.InboxRetention{
		TouchedMailboxes: touchedMailboxes,
		TouchedDomains:   touchedDomains,
	})
	r.mu.Unlock()
	return unseen, nil
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

func (r *Repository) prepareInboxMessage(provider string, mailboxEmail string, input inboxapp.MessageInput, now int64) (*mailboxv1.EmailInboxMessage, string, inboxapp.MessageRow, error) {
	message := input.Message
	mailboxEmail = emailx.Normalize(mailboxEmail)
	if mailboxEmail == "" {
		return nil, "", inboxapp.MessageRow{}, errors.New("mailbox_email is required")
	}
	if message == nil {
		return nil, "", inboxapp.MessageRow{}, errors.New("message is required")
	}
	receivedAt := message.GetReceivedAtUnix()
	if receivedAt <= 0 {
		receivedAt = now
	}
	bodyText := strings.TrimSpace(input.BodyText)
	if bodyText == "" {
		bodyText = strings.TrimSpace(message.GetBodyPreview())
	}
	htmlBody := strings.TrimSpace(input.HTMLBody)
	sourceEmail := emailx.Normalize(stringx.FirstNonEmpty(message.GetSourceMailboxEmail(), message.GetMailboxEmail(), mailboxEmail))
	key := inboxapp.StableMessageKey(provider, mailboxEmail, stringx.FirstNonEmpty(message.GetId(), message.GetSubject(), message.GetBodyPreview()))
	messageID := stringx.FirstNonEmpty(message.GetId(), key)
	bodyPreview := inboxapp.CompactMessageText(message.GetBodyPreview(), 500)
	recipients := inboxapp.UniqueEmails(message.GetRecipients())
	recipientsJSON, err := json.Marshal(recipients)
	if err != nil {
		return nil, "", inboxapp.MessageRow{}, err
	}
	persisted := &mailboxv1.EmailInboxMessage{
		Id:                 messageID,
		MailboxEmail:       mailboxEmail,
		Subject:            strings.TrimSpace(message.GetSubject()),
		FromAddress:        emailx.Normalize(message.GetFromAddress()),
		BodyPreview:        bodyPreview,
		ReceivedAtUnix:     receivedAt,
		Recipients:         recipients,
		ProviderKey:        provider,
		SourceMailboxEmail: sourceEmail,
		BodyArtifactRef:    inboxapp.ArtifactRef(provider, mailboxEmail, messageID, "body_text", int64(len(bodyText)), r.providers.NormalizeProviderInput),
		HtmlArtifactRef:    inboxapp.ArtifactRef(provider, mailboxEmail, messageID, "html_body", int64(len(htmlBody)), r.providers.NormalizeProviderInput),
		RawSize:            message.GetRawSize(),
	}
	row := inboxapp.MessageRow{
		ID:             persisted.GetId(),
		MailboxEmail:   mailboxEmail,
		Subject:        persisted.GetSubject(),
		FromAddress:    persisted.GetFromAddress(),
		BodyPreview:    persisted.GetBodyPreview(),
		ReceivedAtUnix: receivedAt,
		RecipientsJSON: string(recipientsJSON),
		Provider:       provider,
		SourceEmail:    sourceEmail,
		BodyText:       bodyText,
		HTMLBody:       htmlBody,
		RawSize:        persisted.GetRawSize(),
	}
	return persisted, key, row, nil
}

func (r *Repository) trackInboxWatermarkLocked(mailboxEmail string, receivedAtUnix int64, now int64) {
	if receivedAtUnix <= 0 {
		return
	}
	entry, ok := r.mailboxes[emailx.Normalize(mailboxEmail)]
	if !ok || entry.record == nil {
		return
	}
	watermark := time.Unix(receivedAtUnix, 0).UnixNano()
	if watermark > entry.inboxWatermark {
		entry.inboxWatermark = watermark
		entry.record.UpdatedAt = now
		r.mailboxes[emailx.Normalize(mailboxEmail)] = entry
	}
}

func (r *Repository) pruneInboundLocked(provider string, retention mailboxprovider.InboxRetention) {
	definition := r.providers.RetentionByKey(provider)
	if definition == nil {
		return
	}
	policy, ok := definition.RetentionPolicy()
	if !ok {
		return
	}
	switch policy.Scope {
	case mailboxprovider.RetentionScopeDomain:
		for domain := range retention.TouchedDomains {
			r.pruneMessagesLocked(func(message storedMessage) bool {
				return message.row.Provider == provider && domainForEmail(message.row.MailboxEmail) == domain
			}, policy.MaxMessages)
		}
	case mailboxprovider.RetentionScopeMailbox:
		for mailboxEmail := range retention.TouchedMailboxes {
			r.pruneMessagesLocked(func(message storedMessage) bool {
				return message.row.Provider == provider && message.row.MailboxEmail == mailboxEmail
			}, policy.MaxMessages)
		}
	}
}

func (r *Repository) pruneMessagesLocked(match func(storedMessage) bool, keep int) {
	if keep <= 0 {
		return
	}
	matches := []storedMessage{}
	for _, message := range r.messages {
		if match(message) {
			matches = append(matches, message)
		}
	}
	sortStoredMessages(matches)
	for index := keep; index < len(matches); index++ {
		delete(r.messages, messageStorageKey(matches[index].row.Provider, matches[index].row.MailboxEmail, matches[index].key))
	}
}

func (r *Repository) deleteInboxLocked(email string) bool {
	deleted := false
	for key, message := range r.messages {
		if message.row.MailboxEmail == email {
			delete(r.messages, key)
			deleted = true
		}
	}
	return deleted
}

func persistTargetMailboxes(input inboxapp.MessageInput, expandRecipients bool) []string {
	message := input.Message
	if message == nil {
		return []string{}
	}
	if !expandRecipients {
		return inboxapp.UniqueEmails([]string{message.GetMailboxEmail()})
	}
	return inboxapp.MessageMailboxEmails(message.GetMailboxEmail(), message.GetRecipients())
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

func messageStorageKey(provider string, mailboxEmail string, key string) string {
	return strings.Join([]string{
		mailboxprovider.NormalizeKey(provider),
		emailx.Normalize(mailboxEmail),
		strings.TrimSpace(key),
	}, "\x00")
}

func normalizeInboxRowLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	return limit
}
