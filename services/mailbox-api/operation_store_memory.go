package main

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

type memoryOperationStore struct {
	mu   sync.Mutex
	rows map[string]mailboxOperationRow
}

func newMemoryOperationStore() *memoryOperationStore {
	return &memoryOperationStore{rows: map[string]mailboxOperationRow{}}
}

func (s *memoryOperationStore) create(ctx context.Context, operationID, action, emailAddress string) (*mailboxv1.MailboxOperation, error) {
	return s.insert(ctx, mailboxOperationRow{
		OperationID:  strings.TrimSpace(operationID),
		Action:       strings.ToUpper(strings.TrimSpace(action)),
		Status:       operationStatusCreated,
		EmailAddress: emailx.Normalize(emailAddress),
		LastStep:     "created",
	})
}

func (s *memoryOperationStore) createRegistration(ctx context.Context, operationID string, importOnly bool) (*mailboxv1.MailboxOperation, error) {
	return s.insert(ctx, mailboxOperationRow{
		OperationID: strings.TrimSpace(operationID),
		Action:      operationActionRegisterMailbox,
		Status:      operationStatusCreated,
		LastStep:    "queued",
		ImportOnly:  importOnly,
	})
}

func (s *memoryOperationStore) createOAuth(ctx context.Context, operationID string, emailAddress string, onlyMissing bool, limit int32) (*mailboxv1.MailboxOperation, error) {
	return s.insert(ctx, mailboxOperationRow{
		OperationID:  strings.TrimSpace(operationID),
		Action:       operationActionMailboxOAuth,
		Status:       operationStatusCreated,
		EmailAddress: emailx.Normalize(emailAddress),
		LastStep:     "queued",
		OnlyMissing:  onlyMissing,
		Limit:        normalizedLimit(limit),
	})
}

func (s *memoryOperationStore) insert(ctx context.Context, row mailboxOperationRow) (*mailboxv1.MailboxOperation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if row.OperationID == "" {
		return nil, errors.New("operation_id is required")
	}
	now := time.Now().Unix()
	row.CreatedAt = now
	row.UpdatedAt = now
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.rows[row.OperationID]; exists {
		return nil, errors.New("mailbox operation already exists")
	}
	s.rows[row.OperationID] = row
	return operationRowToProto(&row), nil
}

func (s *memoryOperationStore) update(ctx context.Context, operationID string, update operationUpdate) (*mailboxv1.MailboxOperation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	operationID = strings.TrimSpace(operationID)
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[operationID]
	if !ok {
		return nil, errors.New("mailbox operation not found")
	}
	if value := strings.ToUpper(strings.TrimSpace(update.Status)); value != "" {
		row.Status = value
		if value == operationStatusSucceeded || value == operationStatusFailed {
			row.ClaimOwner = ""
			row.ClaimUntil = 0
		}
	}
	if value := strings.TrimSpace(update.LastStep); value != "" {
		row.LastStep = value
	}
	row.ErrorMessage = safeMailboxText(update.ErrorMessage)
	row.ExitCode = update.ExitCode
	row.MailboxCount = update.MailboxCount
	row.FetchedCount = update.FetchedCount
	row.FailedCount = update.FailedCount
	row.MessageCount = update.MessageCount
	row.UpdatedAt = time.Now().Unix()
	s.rows[operationID] = row
	return operationRowToProto(&row), nil
}

func (s *memoryOperationStore) get(ctx context.Context, operationID string) (*mailboxv1.MailboxOperation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	operationID = strings.TrimSpace(operationID)
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[operationID]
	if !ok {
		return nil, errors.New("mailbox operation not found")
	}
	return operationRowToProto(&row), nil
}

func (s *memoryOperationStore) list(ctx context.Context, filter operationListFilter) ([]*mailboxv1.MailboxOperation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	statusValue := operationStatusValue(filter.Status)
	actionValue := operationActionValue(filter.Action)
	emailAddress := emailx.Normalize(filter.EmailAddress)
	s.mu.Lock()
	rows := make([]mailboxOperationRow, 0, len(s.rows))
	for _, row := range s.rows {
		if statusValue != "" && row.Status != statusValue {
			continue
		}
		if actionValue != "" && row.Action != actionValue {
			continue
		}
		if emailAddress != "" && row.EmailAddress != emailAddress {
			continue
		}
		rows = append(rows, row)
	}
	s.mu.Unlock()

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].UpdatedAt > rows[j].UpdatedAt
	})
	if len(rows) > limit {
		rows = rows[:limit]
	}
	operations := make([]*mailboxv1.MailboxOperation, 0, len(rows))
	for i := range rows {
		operations = append(operations, operationRowToProto(&rows[i]))
	}
	return operations, nil
}
