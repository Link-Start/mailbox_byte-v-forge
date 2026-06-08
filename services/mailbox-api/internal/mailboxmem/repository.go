package mailboxmem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
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
	for _, message := range request.Messages {
		for _, mailboxEmail := range persistTargetMailboxes(message, request.ExpandRecipients) {
			touchedMailboxes[mailboxEmail] = struct{}{}
			if domain := domainForEmail(mailboxEmail); domain != "" {
				touchedDomains[domain] = struct{}{}
			}
			persisted, key, row, err := r.prepareInboxMessage(provider, mailboxEmail, message, now)
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

func mergeCredentials(record *mailboxmodel.Record, input *mailboxmodel.Record, existed bool) {
	password := strings.TrimSpace(input.GetPassword())
	refreshToken := strings.TrimSpace(input.GetRefreshToken())
	accessToken := strings.TrimSpace(input.GetAccessToken())
	authStatus := strings.TrimSpace(input.GetAuthStatus())
	lastError := strings.TrimSpace(input.GetLastError())
	if password != "" || !existed {
		record.Password = password
	}
	if refreshToken != "" || !existed {
		record.RefreshToken = refreshToken
	}
	if accessToken != "" || !existed {
		record.AccessToken = accessToken
	}
	switch {
	case authStatus != "":
		record.AuthStatus = authStatus
	case refreshToken != "":
		record.AuthStatus = mailboxmodel.AuthStatusAuthorized
	case !existed && record.AuthStatus == "":
		record.AuthStatus = mailboxmodel.AuthStatusOAuthPending
	}
	if authStatus != "" || lastError != "" || !existed {
		record.LastError = lastError
	}
}

func (r *Repository) listStoredMailboxesLocked(query mailboxprovider.ListQuery) []*mailboxmodel.Record {
	rows := []*mailboxmodel.Record{}
	for _, entry := range r.mailboxes {
		record := cloneRecord(entry.record)
		if !matchesMailboxQuery(record, query) {
			continue
		}
		rows = append(rows, r.project(record))
	}
	return rows
}

func (r *Repository) listVirtualMailboxesLocked(query mailboxprovider.ListQuery) []*mailboxmodel.Record {
	rows := []*mailboxmodel.Record{}
	for _, source := range r.providers.VirtualMailboxSources() {
		if query.Provider != "" && query.Provider != source.Key() {
			continue
		}
		if !source.StoredInboxOnly() || !source.IncludeVirtual(query.AuthStatus) {
			continue
		}
		rows = append(rows, r.virtualMailboxesForProviderLocked(source.Key(), query)...)
	}
	return rows
}

func (r *Repository) virtualMailboxesForProviderLocked(provider string, query mailboxprovider.ListQuery) []*mailboxmodel.Record {
	byEmail := map[string]*mailboxmodel.Record{}
	for _, message := range r.messages {
		if message.row.Provider != provider {
			continue
		}
		email := emailx.Normalize(message.row.MailboxEmail)
		if email == "" || r.mailboxes[email].record != nil {
			continue
		}
		record := byEmail[email]
		if record == nil {
			record = &mailboxmodel.Record{
				EmailAddress: email,
				ProviderKey:  provider,
				CreatedAt:    message.createdAt,
				UpdatedAt:    message.updatedAt,
				Domain:       domainForEmail(email),
			}
			byEmail[email] = record
		}
		if message.createdAt < record.CreatedAt {
			record.CreatedAt = message.createdAt
		}
		if message.updatedAt > record.UpdatedAt {
			record.UpdatedAt = message.updatedAt
		}
	}
	rows := []*mailboxmodel.Record{}
	for _, record := range byEmail {
		if matchesMailboxQuery(record, query) {
			rows = append(rows, r.project(record))
		}
	}
	return rows
}

func matchesMailboxQuery(record *mailboxmodel.Record, query mailboxprovider.ListQuery) bool {
	if record == nil {
		return false
	}
	if query.AuthStatus != "" && strings.TrimSpace(record.GetAuthStatus()) != query.AuthStatus {
		return false
	}
	if query.Provider != "" && mailboxprovider.NormalizeKey(record.GetProviderKey()) != query.Provider {
		return false
	}
	if query.EmailAddress != "" && emailx.Normalize(record.GetEmailAddress()) != query.EmailAddress {
		return false
	}
	if query.HasCursor() {
		updatedAt := record.GetUpdatedAt()
		cursorAt := query.Cursor.UpdatedAt.Unix()
		if updatedAt > cursorAt || (updatedAt == cursorAt && record.GetEmailAddress() >= query.Cursor.ID) {
			return false
		}
	}
	return true
}

func mailboxPageFromRows(rows []*mailboxmodel.Record, limit int) mailboxmodel.ListPage {
	rows = uniqueMailboxRows(rows)
	sortMailboxRows(rows)
	page := pagex.NewKeysetPage(rows, limit, func(mailbox *mailboxmodel.Record) pagex.KeysetCursor {
		return pagex.KeysetCursorValue(time.Unix(mailbox.GetUpdatedAt(), 0).UTC(), mailbox.GetEmailAddress())
	})
	return mailboxmodel.ListPage{Mailboxes: page.Items, NextCursor: page.NextCursor}
}

func uniqueMailboxRows(rows []*mailboxmodel.Record) []*mailboxmodel.Record {
	seen := map[string]struct{}{}
	out := make([]*mailboxmodel.Record, 0, len(rows))
	for _, row := range rows {
		email := emailx.Normalize(row.GetEmailAddress())
		if email == "" {
			continue
		}
		if _, exists := seen[email]; exists {
			continue
		}
		seen[email] = struct{}{}
		out = append(out, row)
	}
	return out
}

func sortMailboxRows(rows []*mailboxmodel.Record) {
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].GetUpdatedAt() == rows[j].GetUpdatedAt() {
			return rows[i].GetEmailAddress() > rows[j].GetEmailAddress()
		}
		return rows[i].GetUpdatedAt() > rows[j].GetUpdatedAt()
	})
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

func (r *Repository) prepareInboxMessage(provider string, mailboxEmail string, message *mailboxv1.EmailInboxMessage, now int64) (*mailboxv1.EmailInboxMessage, string, inboxapp.MessageRow, error) {
	mailboxEmail = emailx.Normalize(mailboxEmail)
	if mailboxEmail == "" {
		return nil, "", inboxapp.MessageRow{}, errors.New("mailbox_email is required")
	}
	receivedAt := message.GetReceivedAtUnix()
	if receivedAt <= 0 {
		receivedAt = now
	}
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
		BodyArtifactRef:    inboxapp.ArtifactRef(provider, mailboxEmail, messageID, "body_text", int64(len(message.GetBodyPreview())), r.providers.NormalizeProviderInput),
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
		BodyText:       strings.TrimSpace(message.GetBodyPreview()),
		HTMLBody:       "",
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

func persistTargetMailboxes(message *mailboxv1.EmailInboxMessage, expandRecipients bool) []string {
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
