package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"

	"mailboxapi/pb"
)

func (a *mailboxActivities) runMailboxRegistration(ctx context.Context, input mailboxRegistrationActionInput) (mailboxOperationResult, error) {
	operationID := strings.TrimSpace(input.OperationID)
	if err := a.markRunning(ctx, operationID, "run_registration"); err != nil {
		return mailboxOperationResult{OperationID: operationID, ErrorMessage: safeMailboxError(err)}, err
	}

	resp, err := a.outlookRegistration.RunMailboxRegistration(ctx, &pb.RunMailboxRegistrationRequest{
		Enabled:    !input.ImportOnly,
		ImportOnly: input.ImportOnly,
	})
	if err != nil {
		message := safeMailboxErrorMessage("run mailbox registration", err)
		a.updateOperation(ctx, operationID, operationUpdate{
			Status:       operationStatusFailed,
			LastStep:     "run_registration",
			ErrorMessage: message,
		})
		return mailboxOperationResult{OperationID: operationID, Success: false, ErrorMessage: message}, err
	}
	if resp == nil {
		message := "mailbox registration runner returned empty response"
		a.updateOperation(ctx, operationID, operationUpdate{
			Status:       operationStatusFailed,
			LastStep:     "run_registration",
			ErrorMessage: message,
		})
		return mailboxOperationResult{OperationID: operationID, Success: false, ErrorMessage: message}, fmt.Errorf("%s", message)
	}

	result := mailboxOperationResult{
		OperationID:  operationID,
		Success:      resp.GetSuccess(),
		ExitCode:     resp.GetExitCode(),
		ErrorMessage: strings.TrimSpace(resp.GetErrorMessage()),
		MailboxCount: registeredMailboxCount(resp.GetAccounts()),
	}
	if result.Success {
		if err := a.persistRegisteredAccounts(ctx, resp.GetAccounts()); err != nil {
			result.Success = false
			result.ErrorMessage = safeMailboxErrorMessage("persist registered mailboxes", err)
		}
	}
	if !result.Success && result.ErrorMessage == "" {
		result.ErrorMessage = "mailbox registration failed"
	}
	a.finishOperation(ctx, operationID, "run_registration", result)
	return result, nil
}

func registeredMailboxCount(accounts []*pb.MailboxRegistrationAccount) int32 {
	var count int32
	for _, account := range accounts {
		if emailx.Normalize(account.GetEmailAddress()) != "" {
			count++
		}
	}
	return count
}

func (a *mailboxActivities) persistRegisteredAccounts(ctx context.Context, accounts []*pb.MailboxRegistrationAccount) error {
	for _, account := range accounts {
		email := emailx.Normalize(account.GetEmailAddress())
		password := strings.TrimSpace(account.GetPassword())
		if email == "" {
			continue
		}
		if password == "" {
			return fmt.Errorf("mailbox account missing password: %s", emailx.Redact(email))
		}
		if err := a.upsertMailbox(ctx, &pb.EmailMailbox{
			EmailAddress: email,
			Password:     password,
			RefreshToken: strings.TrimSpace(account.GetRefreshToken()),
			AccessToken:  strings.TrimSpace(account.GetAccessToken()),
			AuthStatus:   mailboxAuthStatus(account.GetRefreshToken(), ""),
			LastError:    "",
		}); err != nil {
			return err
		}
	}
	return nil
}
