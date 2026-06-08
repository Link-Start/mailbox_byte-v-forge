package main

import "mailboxapi/pb"

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
