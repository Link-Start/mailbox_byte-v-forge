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

type inboxFetchTarget struct {
	fetchMailbox  *mailboxmodel.Record
	resultMailbox *mailboxmodel.Record
}

func (s *EmailService) inboxFetchTargets(ctx context.Context, request *mailboxv1.FetchMailboxInboxesRequest) ([]inboxFetchTarget, *mailboxv1.FetchMailboxInboxesResponse, error) {
	targets := []inboxFetchTarget{}
	requestedEmail := emailx.Normalize(request.GetEmailAddress())
	if requestedEmail == "" {
		mailboxes, err := s.mailboxRepo.ListOAuthMailboxes(ctx, request.GetMaxMailboxes())
		if err != nil {
			return nil, nil, status.Error(codes.Internal, safeMailboxError(err))
		}
		for _, mailbox := range mailboxes {
			targets = append(targets, inboxFetchTarget{fetchMailbox: mailbox, resultMailbox: mailbox})
		}
		return targets, nil, nil
	}

	if response, ok, err := s.storedOnlyInboxResponse(ctx, requestedEmail, request); ok || err != nil {
		return nil, response, err
	}
	fetchMailbox, err := s.mailboxRepo.PollMailboxForEmail(ctx, requestedEmail)
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, safeMailboxError(err))
	}
	resultMailbox := fetchMailbox
	if mailbox, err := s.mailboxRepo.FindMailbox(ctx, requestedEmail); err == nil {
		resultMailbox = mailbox
	}
	return append(targets, inboxFetchTarget{fetchMailbox: fetchMailbox, resultMailbox: resultMailbox}), nil, nil
}

func (s *EmailService) storedOnlyInboxResponse(ctx context.Context, email string, request *mailboxv1.FetchMailboxInboxesRequest) (*mailboxv1.FetchMailboxInboxesResponse, bool, error) {
	resultMailbox, ok := s.providers.StoredInboxOnlyMailbox(email)
	if !ok {
		return nil, false, nil
	}
	if mailbox, err := s.mailboxRepo.FindMailbox(ctx, email); err == nil {
		resultMailbox = mailbox
	}
	messages, err := s.inbox.ListMessagesSince(ctx, email, request.GetLimitPerMailbox(), request.GetReceivedAfterUnix())
	if err != nil {
		return nil, true, status.Error(codes.Internal, safeMailboxError(err))
	}
	return &mailboxv1.FetchMailboxInboxesResponse{
		MailboxCount: int32(1),
		FetchedCount: int32(1),
		MessageCount: int32(len(messages)),
		Results: []*mailboxv1.FetchMailboxInboxResult{{
			Mailbox:  mailboxapp.PublicMailbox(resultMailbox),
			Messages: messages,
		}},
	}, true, nil
}
