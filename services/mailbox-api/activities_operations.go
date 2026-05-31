package main

import (
	"context"
	"log"
	"strings"
)

func (a *mailboxActivities) markRunning(ctx context.Context, operationID string, step string) error {
	_, err := a.operations.update(ctx, operationID, operationUpdate{
		Status:   operationStatusRunning,
		LastStep: step,
	})
	return err
}

func (a *mailboxActivities) finishOperation(ctx context.Context, operationID string, step string, result mailboxOperationResult) {
	statusValue := operationStatusSucceeded
	if !result.Success || strings.TrimSpace(result.ErrorMessage) != "" {
		statusValue = operationStatusFailed
	}
	a.updateOperation(ctx, operationID, operationUpdate{
		Status:       statusValue,
		LastStep:     step,
		ErrorMessage: result.ErrorMessage,
		ExitCode:     result.ExitCode,
		MailboxCount: result.MailboxCount,
		FetchedCount: result.FetchedCount,
		FailedCount:  result.FailedCount,
		MessageCount: result.MessageCount,
	})
}

func (a *mailboxActivities) updateOperation(ctx context.Context, operationID string, update operationUpdate) {
	operation, err := a.operations.update(ctx, operationID, update)
	if err != nil {
		log.Printf("update mailbox operation failed operation=%s: %s", operationID, safeMailboxError(err))
		return
	}
	a.hot.PublishOperation(ctx, operation)
}
