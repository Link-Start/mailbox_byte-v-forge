package main

import (
	"context"
	"fmt"
	"strings"

	"mailboxapi/internal/emailx"

	"mailboxapi/pb"
)

func (a *mailboxActivities) SelectMailboxOAuthAccounts(ctx context.Context, req *pb.SelectMailboxOAuthAccountsRequest) (*pb.SelectMailboxOAuthAccountsResponse, error) {
	operationID := strings.TrimSpace(req.GetOperationId())
	if err := a.markRunning(ctx, operationID, "run_oauth"); err != nil {
		return nil, err
	}

	accounts, err := a.oauthAccounts(ctx, req.GetEmailAddress(), req.GetOnlyMissing(), normalizedLimit(req.GetLimit()))
	if err != nil {
		message := safeMailboxErrorMessage("select mailbox OAuth accounts", err)
		a.updateOperation(ctx, operationID, operationUpdate{
			Status:       operationStatusFailed,
			LastStep:     "run_oauth",
			ErrorMessage: message,
		})
		return &pb.SelectMailboxOAuthAccountsResponse{}, nil
	}
	return &pb.SelectMailboxOAuthAccountsResponse{Accounts: accounts}, nil
}

func (a *mailboxActivities) RunMailboxOAuthAccount(ctx context.Context, req *pb.RunMailboxOAuthAccountRequest) (*pb.RunMailboxOAuthAccountResponse, error) {
	account := req.GetAccount()
	if account == nil || emailx.Normalize(account.GetEmailAddress()) == "" {
		return &pb.RunMailboxOAuthAccountResponse{Result: &pb.MailboxOAuthResult{
			Success:      false,
			ErrorMessage: "mailbox OAuth account is required",
		}}, nil
	}
	resp, err := a.providerActions.RunOAuth(ctx, "", &pb.RunMailboxOAuthRequest{
		EmailAddress: emailx.Normalize(account.GetEmailAddress()),
		OnlyMissing:  false,
		Limit:        1,
		Accounts:     []*pb.MailboxRegistrationAccount{account},
	})
	if err != nil {
		message := safeMailboxErrorMessage("run mailbox OAuth", err)
		return nil, fmt.Errorf("%s", message)
	}
	if resp == nil {
		message := "mailbox registration runner returned empty OAuth response"
		return nil, fmt.Errorf("%s", message)
	}
	if len(resp.GetResults()) == 0 {
		return &pb.RunMailboxOAuthAccountResponse{Result: &pb.MailboxOAuthResult{
			EmailAddress: emailx.Normalize(account.GetEmailAddress()),
			Success:      false,
			ErrorMessage: "mailbox OAuth returned no result",
		}}, nil
	}
	return &pb.RunMailboxOAuthAccountResponse{Result: resp.GetResults()[0]}, nil
}

func (a *mailboxActivities) CompleteMailboxOAuth(ctx context.Context, req *pb.CompleteMailboxOAuthRequest) (mailboxOperationResult, error) {
	operationID := strings.TrimSpace(req.GetOperationId())
	response := mailboxOAuthResponse(req.GetAccounts(), req.GetResults())
	result := mailboxOperationResult{
		OperationID:  operationID,
		Success:      response.GetSuccess(),
		ErrorMessage: strings.TrimSpace(response.GetErrorMessage()),
		MailboxCount: response.GetProcessed(),
		FetchedCount: response.GetSucceeded(),
		FailedCount:  response.GetFailed(),
	}
	if err := a.persistOAuthResults(ctx, response.GetResults(), req.GetAccounts()); err != nil {
		result.Success = false
		result.ErrorMessage = safeMailboxErrorMessage("persist mailbox OAuth results", err)
	}
	if !result.Success && result.ErrorMessage == "" {
		result.ErrorMessage = "mailbox OAuth failed"
	}
	if len(req.GetAccounts()) == 0 {
		result.ErrorMessage = "no mailbox accounts eligible for OAuth"
		a.updateOperation(ctx, operationID, operationUpdate{
			Status:       operationStatusFailed,
			LastStep:     "run_oauth",
			ErrorMessage: result.ErrorMessage,
		})
		return result, nil
	}
	a.finishOperation(ctx, operationID, "run_oauth", result)
	return result, nil
}

func mailboxOAuthResponse(accounts []*pb.MailboxRegistrationAccount, results []*pb.MailboxOAuthResult) *pb.RunMailboxOAuthResponse {
	response := &pb.RunMailboxOAuthResponse{Processed: int32(len(accounts)), Results: results}
	for _, result := range results {
		if result.GetSuccess() {
			response.Succeeded++
		} else {
			response.Failed++
		}
	}
	response.Success = response.Processed > 0 && response.Failed == 0 && response.Succeeded > 0
	if !response.Success {
		response.ErrorMessage = fmt.Sprintf("mailbox OAuth failed: %d/%d", response.Failed, response.Processed)
	}
	return response
}
