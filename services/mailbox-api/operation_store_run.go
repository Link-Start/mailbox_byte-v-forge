package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/dbclaim"
	"gorm.io/gorm"
)

func (s *operationStore) startRegistrationWorkerRun(ctx context.Context, operationID string) (*operationRunStart, error) {
	return s.startWorkerRun(ctx, operationID, operationActionRegisterMailbox, "run_registration", "mailbox-registration-worker")
}

func (s *operationStore) startOAuthWorkerRun(ctx context.Context, operationID string) (*operationRunStart, error) {
	return s.startWorkerRun(ctx, operationID, operationActionMailboxOAuth, "run_oauth", "mailbox-oauth-worker")
}

func (s *operationStore) startWorkerRun(ctx context.Context, operationID string, action string, runStep string, workerID string) (*operationRunStart, error) {
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return nil, errors.New("operation_id is required")
	}
	now := time.Now().Unix()
	runLeaseUntil := dbclaim.Until(now, operationActionRunLeaseSeconds)
	action = strings.ToUpper(strings.TrimSpace(action))
	workerID = strings.TrimSpace(workerID)

	var row mailboxOperationRow
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(dbclaim.ForUpdate()).First(&row, "operation_id = ?", operationID).Error; err != nil {
			return err
		}
		if row.Action != action {
			return errOperationInvalidAction
		}
		switch row.Status {
		case operationStatusSucceeded, operationStatusFailed:
			return nil
		case operationStatusRunning:
			if row.LastStep == runStep && row.ClaimUntil > now {
				return errOperationAlreadyRunning
			}
		}
		if err := tx.Model(&mailboxOperationRow{}).
			Where("operation_id = ?", operationID).
			Updates(dbclaim.ClaimUpdates(operationStatusRunning, runStep, "", workerID, runLeaseUntil)).Error; err != nil {
			return err
		}
		return tx.First(&row, "operation_id = ?", operationID).Error
	})
	if err != nil {
		return nil, err
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
