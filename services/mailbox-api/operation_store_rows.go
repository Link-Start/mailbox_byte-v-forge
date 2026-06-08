package main

import (
	"fmt"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

type operationRowScanner interface {
	Scan(dest ...any) error
}

func scanOperationRow(scanner operationRowScanner) (mailboxOperationRow, error) {
	var row mailboxOperationRow
	err := scanner.Scan(
		&row.OperationID,
		&row.Action,
		&row.Status,
		&row.EmailAddress,
		&row.LastStep,
		&row.ErrorMessage,
		&row.ImportOnly,
		&row.OnlyMissing,
		&row.Limit,
		&row.ClaimOwner,
		&row.ClaimUntil,
		&row.AttemptCount,
		&row.ExitCode,
		&row.MailboxCount,
		&row.FetchedCount,
		&row.FailedCount,
		&row.MessageCount,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	return row, err
}

func operationSelectSQL() string {
	return fmt.Sprintf(`SELECT %s FROM mailbox_operations`, operationColumns())
}

func operationColumns() string {
	return `operation_id, action, status, email_address, last_step, error_message,
		import_only, only_missing, "limit", claim_owner, claim_until, attempt_count,
		exit_code, mailbox_count, fetched_count, failed_count, message_count, created_at, updated_at`
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

func operationRunStartFromRow(row *mailboxOperationRow) *operationRunStart {
	if row == nil {
		return nil
	}
	return &operationRunStart{
		Operation:    operationRowToProto(row),
		EmailAddress: row.EmailAddress,
		ImportOnly:   row.ImportOnly,
		OnlyMissing:  row.OnlyMissing,
		Limit:        row.Limit,
		Final:        operationRowIsFinal(row),
	}
}

func operationRowIsFinal(row *mailboxOperationRow) bool {
	return row != nil && (row.Status == operationStatusSucceeded || row.Status == operationStatusFailed)
}
