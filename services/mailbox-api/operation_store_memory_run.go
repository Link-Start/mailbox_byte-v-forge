package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"mailboxapi/internal/dbclaim"
)

func (s *memoryOperationStore) startRegistrationWorkerRun(ctx context.Context, operationID string) (*operationRunStart, error) {
	return s.startWorkerRun(ctx, operationID, operationActionRegisterMailbox, "run_registration", "mailbox-registration-worker")
}

func (s *memoryOperationStore) startOAuthWorkerRun(ctx context.Context, operationID string) (*operationRunStart, error) {
	return s.startWorkerRun(ctx, operationID, operationActionMailboxOAuth, "run_oauth", "mailbox-oauth-worker")
}

func (s *memoryOperationStore) startWorkerRun(ctx context.Context, operationID string, action string, runStep string, workerID string) (*operationRunStart, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return nil, errors.New("operation_id is required")
	}
	now := time.Now().Unix()
	action = strings.ToUpper(strings.TrimSpace(action))
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[operationID]
	if !ok {
		return nil, errors.New("mailbox operation not found")
	}
	if row.Action != action {
		return nil, errOperationInvalidAction
	}
	if row.Status != operationStatusSucceeded && row.Status != operationStatusFailed {
		if row.Status == operationStatusRunning && row.LastStep == runStep && row.ClaimUntil > now {
			return nil, errOperationAlreadyRunning
		}
		row.Status = operationStatusRunning
		row.LastStep = runStep
		row.ErrorMessage = ""
		row.ClaimOwner = strings.TrimSpace(workerID)
		row.ClaimUntil = dbclaim.Until(now, operationActionRunLeaseSeconds)
		row.AttemptCount++
		row.UpdatedAt = now
		s.rows[operationID] = row
	}
	return &operationRunStart{
		Operation:    operationRowToProto(&row),
		EmailAddress: row.EmailAddress,
		ImportOnly:   row.ImportOnly,
		OnlyMissing:  row.OnlyMissing,
		Limit:        row.Limit,
		Final:        row.Status == operationStatusSucceeded || row.Status == operationStatusFailed,
	}, nil
}
