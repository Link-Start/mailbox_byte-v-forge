package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/byte-v-forge/common-lib/envx"

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
	if !req.GetEnabled() || !envx.Bool("OUTLOOK_REGISTER_ENABLED", false) {
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

func registrationResponse(records []mailboxRecord, err error) *pb.RunMailboxRegistrationResponse {
	if err != nil {
		return &pb.RunMailboxRegistrationResponse{Success: false, ExitCode: 1, ErrorMessage: safeMailboxError(err)}
	}
	accounts := make([]*pb.MailboxRegistrationAccount, 0, len(records))
	for _, record := range records {
		accounts = append(accounts, &pb.MailboxRegistrationAccount{
			EmailAddress: record.email,
			Password:     record.password,
			RefreshToken: record.refreshToken,
			AccessToken:  record.accessToken,
			Source:       record.source,
		})
	}
	errorMessage := ""
	if len(accounts) == 0 {
		errorMessage = "no mailbox records found to import"
	}
	return &pb.RunMailboxRegistrationResponse{
		Success:      len(accounts) > 0,
		ExitCode:     0,
		ErrorMessage: errorMessage,
		Accounts:     accounts,
	}
}

func selectOAuthTargets(req *pb.RunMailboxOAuthRequest) []*pb.MailboxRegistrationAccount {
	requested := emailx.Normalize(req.GetEmailAddress())
	limit := req.GetLimit()
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	targets := make([]*pb.MailboxRegistrationAccount, 0, len(req.GetAccounts()))
	for _, account := range req.GetAccounts() {
		email := emailx.Normalize(account.GetEmailAddress())
		if email == "" {
			continue
		}
		if requested != "" && email != requested {
			continue
		}
		if req.GetOnlyMissing() && strings.TrimSpace(account.GetRefreshToken()) != "" {
			continue
		}
		targets = append(targets, &pb.MailboxRegistrationAccount{
			EmailAddress: email,
			Password:     strings.TrimSpace(account.GetPassword()),
			RefreshToken: strings.TrimSpace(account.GetRefreshToken()),
			AccessToken:  strings.TrimSpace(account.GetAccessToken()),
			Source:       strings.TrimSpace(account.GetSource()),
		})
		if requested == "" && int32(len(targets)) >= limit {
			break
		}
	}
	return targets
}
