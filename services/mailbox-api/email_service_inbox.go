package main

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"

	"mailboxapi/internal/mailboxapp"
	"mailboxapi/internal/mailboxmodel"
)

func (s *EmailService) ListInbox(ctx context.Context, request *mailboxv1.ListMailboxInboxRequest) (*mailboxv1.ListMailboxInboxResponse, error) {
	email := emailx.Normalize(request.GetEmailAddress())
	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "email_address is required")
	}
	limit := request.GetLimit()
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	messages, err := s.inbox.ListMessages(ctx, email, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, safeMailboxError(err))
	}
	resultMailbox := &mailboxmodel.Record{
		EmailAddress: email,
		ProviderKey:  s.providers.ProviderForInboxAddress(email, messages),
		Domain:       domainForEmail(email),
	}
	s.providers.prepareProjection(resultMailbox)
	if mailbox, err := s.mailboxRepo.FindMailbox(ctx, email); err == nil {
		resultMailbox = mailbox
	}
	return &mailboxv1.ListMailboxInboxResponse{Result: &mailboxv1.FetchMailboxInboxResult{
		Mailbox:  mailboxapp.PublicMailbox(resultMailbox),
		Messages: messages,
	}}, nil
}
