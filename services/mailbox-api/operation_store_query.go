package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

func (s *pgOperationStore) update(ctx context.Context, operationID string, update operationUpdate) (*mailboxv1.MailboxOperation, error) {
	status := strings.ToUpper(strings.TrimSpace(update.Status))
	lastStep := strings.TrimSpace(update.LastStep)
	clearClaim := status == operationStatusSucceeded || status == operationStatusFailed
	query := `UPDATE mailbox_operations SET
		status = CASE WHEN $2 <> '' THEN $2 ELSE status END,
		last_step = CASE WHEN $3 <> '' THEN $3 ELSE last_step END,
		error_message = $4,
		exit_code = $5,
		mailbox_count = $6,
		fetched_count = $7,
		failed_count = $8,
		message_count = $9,
		claim_owner = CASE WHEN $10 THEN '' ELSE claim_owner END,
		claim_until = CASE WHEN $10 THEN 0 ELSE claim_until END,
		updated_at = $11
		WHERE operation_id = $1
		RETURNING ` + operationColumns()
	row, err := scanOperationRow(s.pool.QueryRow(ctx, query,
		strings.TrimSpace(operationID),
		status,
		lastStep,
		safeMailboxText(update.ErrorMessage),
		update.ExitCode,
		update.MailboxCount,
		update.FetchedCount,
		update.FailedCount,
		update.MessageCount,
		clearClaim,
		time.Now().Unix(),
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errOperationNotFound
	}
	if err != nil {
		return nil, err
	}
	return operationRowToProto(&row), nil
}

func (s *pgOperationStore) get(ctx context.Context, operationID string) (*mailboxv1.MailboxOperation, error) {
	row, err := scanOperationRow(s.pool.QueryRow(ctx, operationSelectSQL()+` WHERE operation_id = $1`, strings.TrimSpace(operationID)))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errOperationNotFound
	}
	if err != nil {
		return nil, err
	}
	return operationRowToProto(&row), nil
}

func (s *pgOperationStore) list(ctx context.Context, filter operationListFilter) ([]*mailboxv1.MailboxOperation, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	conditions := []string{}
	args := []any{}
	if value := operationStatusValue(filter.Status); value != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", appendOperationArg(&args, value)))
	}
	if value := operationActionValue(filter.Action); value != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", appendOperationArg(&args, value)))
	}
	if value := emailx.Normalize(filter.EmailAddress); value != "" {
		conditions = append(conditions, fmt.Sprintf("email_address = $%d", appendOperationArg(&args, value)))
	}
	query := operationSelectSQL()
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	args = append(args, limit)
	query += fmt.Sprintf(" ORDER BY updated_at DESC LIMIT $%d", len(args))
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	operations := []*mailboxv1.MailboxOperation{}
	for rows.Next() {
		row, err := scanOperationRow(rows)
		if err != nil {
			return nil, err
		}
		operations = append(operations, operationRowToProto(&row))
	}
	return operations, rows.Err()
}

func appendOperationArg(args *[]any, value any) int {
	*args = append(*args, value)
	return len(*args)
}

func operationRowToProto(row *mailboxOperationRow) *mailboxv1.MailboxOperation {
	if row == nil {
		return nil
	}
	return &mailboxv1.MailboxOperation{
		OperationId:  row.OperationID,
		Action:       publicOperationAction(row.Action),
		Status:       publicOperationStatus(row.Status),
		EmailAddress: row.EmailAddress,
		LastStep:     row.LastStep,
		ErrorMessage: row.ErrorMessage,
		ExitCode:     row.ExitCode,
		MailboxCount: row.MailboxCount,
		FetchedCount: row.FetchedCount,
		FailedCount:  row.FailedCount,
		MessageCount: row.MessageCount,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
