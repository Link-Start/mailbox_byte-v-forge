package main

import (
	"context"
	"fmt"
	"strings"

	"mailboxapi/pb"
)

func (r *outlookRegistrationRunner) RunMailboxRegistration(ctx context.Context, req *pb.RunMailboxRegistrationRequest) (*pb.RunMailboxRegistrationResponse, error) {
	records, err := readMailboxRecords(r.cfg.resultsDir, true)
	if req.GetImportOnly() {
		return registrationResponse(records, err), nil
	}
	if err != nil {
		return &pb.RunMailboxRegistrationResponse{Success: false, ExitCode: 1, ErrorMessage: safeMailboxError(err)}, nil
	}
	if len(records) > 0 {
		return registrationResponse(records, nil), nil
	}
	if !req.GetEnabled() || !r.cfg.enabled {
		return &pb.RunMailboxRegistrationResponse{Success: false, ExitCode: 0, ErrorMessage: "mailbox registration is disabled"}, nil
	}
	return &pb.RunMailboxRegistrationResponse{
		Success:      false,
		ExitCode:     1,
		ErrorMessage: "Outlook account creation is modeled in mailbox-api and must execute through browser-automation before enabling this action",
	}, nil
}

func (r *outlookRegistrationRunner) RunMailboxOAuth(ctx context.Context, req *pb.RunMailboxOAuthRequest) (*pb.RunMailboxOAuthResponse, error) {
	targets := selectOAuthTargets(req)
	if len(targets) == 0 {
		return &pb.RunMailboxOAuthResponse{
			Success:      false,
			Processed:    0,
			ErrorMessage: "mailbox OAuth accounts are required",
		}, nil
	}

	response := &pb.RunMailboxOAuthResponse{Processed: int32(len(targets))}
	for _, account := range targets {
		result := &pb.MailboxOAuthResult{EmailAddress: account.GetEmailAddress()}
		if strings.TrimSpace(account.GetPassword()) == "" {
			result.ErrorMessage = "mailbox password is required for OAuth"
			response.Failed++
			response.Results = append(response.Results, result)
			continue
		}
		tokens, err := r.runBrowserOAuth(ctx, account.GetEmailAddress(), account.GetPassword())
		if err != nil {
			result.ErrorMessage = safeMailboxError(err)
			response.Failed++
			response.Results = append(response.Results, result)
			continue
		}
		result.Success = true
		result.RefreshToken = tokens.refreshToken
		result.AccessToken = tokens.accessToken
		response.Succeeded++
		response.Results = append(response.Results, result)
	}
	response.Success = response.Failed == 0 && response.Succeeded > 0
	if !response.Success {
		response.ErrorMessage = fmt.Sprintf("mailbox OAuth failed: %d/%d", response.Failed, response.Processed)
	}
	return response, nil
}
