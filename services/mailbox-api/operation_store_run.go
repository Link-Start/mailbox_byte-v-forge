package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/dbclaim"
)

func (s *pgOperationStore) startRegistrationWorkerRun(ctx context.Context, operationID string) (*operationRunStart, error) {
	return s.startWorkerRun(ctx, operationID, operationActionRegisterMailbox, "run_registration", "mailbox-registration-worker")
}

func (s *pgOperationStore) startOAuthWorkerRun(ctx context.Context, operationID string) (*operationRunStart, error) {
	return s.startWorkerRun(ctx, operationID, operationActionMailboxOAuth, "run_oauth", "mailbox-oauth-worker")
}

func (s *pgOperationStore) startWorkerRun(ctx context.Context, operationID string, action string, runStep string, workerID string) (*operationRunStart, error) {
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return nil, errors.New("operation_id is required")
	}
	now := time.Now().Unix()
	runLeaseUntil := dbclaim.Until(now, operationActionRunLeaseSeconds)
	action = strings.ToUpper(strings.TrimSpace(action))
	workerID = strings.TrimSpace(workerID)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	row, err := scanOperationRow(tx.QueryRow(ctx, operationSelectSQL()+` WHERE operation_id = $1 FOR UPDATE`, operationID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errOperationNotFound
	}
	if err != nil {
		return nil, err
	}
	if row.Action != action {
		return nil, errOperationInvalidAction
	}
	switch row.Status {
	case operationStatusSucceeded, operationStatusFailed:
		// final operations are safe to ack without re-running
	case operationStatusRunning:
		if row.LastStep == runStep && row.ClaimUntil > now {
			return nil, errOperationAlreadyRunning
		}
		row, err = s.claimWorkerRun(ctx, tx, operationID, runStep, workerID, runLeaseUntil, now)
	default:
		row, err = s.claimWorkerRun(ctx, tx, operationID, runStep, workerID, runLeaseUntil, now)
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true
	return &operationRunStart{
		Operation:    operationRowToProto(&row),
		EmailAddress: row.EmailAddress,
		ImportOnly:   row.ImportOnly,
		OnlyMissing:  row.OnlyMissing,
		Limit:        row.Limit,
		Final:        row.Status == operationStatusSucceeded || row.Status == operationStatusFailed,
	}, nil
}

func (s *pgOperationStore) claimWorkerRun(ctx context.Context, tx pgx.Tx, operationID string, runStep string, workerID string, claimUntil int64, now int64) (mailboxOperationRow, error) {
	query := `UPDATE mailbox_operations SET
		status = $2,
		last_step = $3,
		error_message = '',
		claim_owner = $4,
		claim_until = $5,
		attempt_count = attempt_count + 1,
		updated_at = $6
		WHERE operation_id = $1
		RETURNING ` + operationColumns()
	return scanOperationRow(tx.QueryRow(ctx, query, operationID, operationStatusRunning, runStep, workerID, claimUntil, now))
}
