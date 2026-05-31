package main

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/dbclaim"
	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
)

func (s *operationStore) update(ctx context.Context, operationID string, update operationUpdate) (*mailboxv1.MailboxOperation, error) {
	updates := map[string]any{}
	if value := strings.ToUpper(strings.TrimSpace(update.Status)); value != "" {
		updates["status"] = value
		if value == operationStatusSucceeded || value == operationStatusFailed {
			for key, item := range dbclaim.ClearUpdates() {
				updates[key] = item
			}
		}
	}
	if value := strings.TrimSpace(update.LastStep); value != "" {
		updates["last_step"] = value
	}
	updates["error_message"] = safeMailboxText(update.ErrorMessage)
	updates["exit_code"] = update.ExitCode
	updates["mailbox_count"] = update.MailboxCount
	updates["fetched_count"] = update.FetchedCount
	updates["failed_count"] = update.FailedCount
	updates["message_count"] = update.MessageCount

	if err := s.db.WithContext(ctx).Model(&mailboxOperationRow{}).
		Where("operation_id = ?", strings.TrimSpace(operationID)).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.get(ctx, operationID)
}

func (s *operationStore) get(ctx context.Context, operationID string) (*mailboxv1.MailboxOperation, error) {
	var row mailboxOperationRow
	if err := s.db.WithContext(ctx).First(&row, "operation_id = ?", strings.TrimSpace(operationID)).Error; err != nil {
		return nil, err
	}
	return operationRowToProto(&row), nil
}

func (s *operationStore) list(ctx context.Context, filter operationListFilter) ([]*mailboxv1.MailboxOperation, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := s.db.WithContext(ctx).Model(&mailboxOperationRow{})
	if value := strings.ToUpper(strings.TrimSpace(filter.Status)); value != "" {
		query = query.Where("status = ?", value)
	}
	if value := strings.ToUpper(strings.TrimSpace(filter.Action)); value != "" {
		query = query.Where("action = ?", value)
	}
	if value := emailx.Normalize(filter.EmailAddress); value != "" {
		query = query.Where("email_address = ?", value)
	}

	var rows []mailboxOperationRow
	if err := query.Order("updated_at DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	operations := make([]*mailboxv1.MailboxOperation, 0, len(rows))
	for i := range rows {
		operations = append(operations, operationRowToProto(&rows[i]))
	}
	return operations, nil
}

func operationRowToProto(row *mailboxOperationRow) *mailboxv1.MailboxOperation {
	if row == nil {
		return nil
	}
	return &mailboxv1.MailboxOperation{
		OperationId:  row.OperationID,
		Action:       row.Action,
		Status:       row.Status,
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
