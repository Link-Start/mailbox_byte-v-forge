package main

import (
	"context"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mailboxapi/pb"
)

type inboxFetchTarget struct {
	fetchMailbox  *pb.EmailMailbox
	resultMailbox *pb.EmailMailbox
}

func (s *EmailService) FetchInboxes(ctx context.Context, request *mailboxv1.FetchMailboxInboxesRequest) (*mailboxv1.FetchMailboxInboxesResponse, error) {
	unlock, err := s.acquireInboxLock(ctx)
	if err != nil {
		return nil, err
	}
	defer unlock()

	targets, response, err := s.inboxFetchTargets(ctx, request)
	if err != nil || response != nil {
		return response, err
	}
	return s.fetchInboxTargets(ctx, request, targets)
}

func (s *EmailService) inboxFetchTargets(ctx context.Context, request *mailboxv1.FetchMailboxInboxesRequest) ([]inboxFetchTarget, *mailboxv1.FetchMailboxInboxesResponse, error) {
	targets := []inboxFetchTarget{}
	requestedEmail := emailx.Normalize(request.GetEmailAddress())
	if requestedEmail == "" {
		mailboxes, err := s.store.ListOAuthMailboxes(ctx, request.GetMaxMailboxes())
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
	fetchMailbox, err := s.store.PollMailboxForEmail(ctx, requestedEmail)
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, safeMailboxError(err))
	}
	resultMailbox := fetchMailbox
	if mailbox, err := s.store.FindMailbox(ctx, requestedEmail); err == nil {
		resultMailbox = mailbox
	}
	return append(targets, inboxFetchTarget{fetchMailbox: fetchMailbox, resultMailbox: resultMailbox}), nil, nil
}

func (s *EmailService) storedOnlyInboxResponse(ctx context.Context, email string, request *mailboxv1.FetchMailboxInboxesRequest) (*mailboxv1.FetchMailboxInboxesResponse, bool, error) {
	resultMailbox, ok := s.providers.StoredInboxOnlyMailbox(email)
	if !ok {
		return nil, false, nil
	}
	if mailbox, err := s.store.FindMailbox(ctx, email); err == nil {
		resultMailbox = mailbox
	}
	messages, err := s.store.ListInboxMessagesSince(ctx, email, request.GetLimitPerMailbox(), request.GetReceivedAfterUnix())
	if err != nil {
		return nil, true, status.Error(codes.Internal, safeMailboxError(err))
	}
	return &mailboxv1.FetchMailboxInboxesResponse{
		MailboxCount: int32(1),
		FetchedCount: int32(1),
		MessageCount: int32(len(messages)),
		Results: []*mailboxv1.FetchMailboxInboxResult{{
			Mailbox:  publicMailbox(resultMailbox),
			Messages: messages,
		}},
	}, true, nil
}

func (s *EmailService) fetchInboxTargets(ctx context.Context, request *mailboxv1.FetchMailboxInboxesRequest, targets []inboxFetchTarget) (*mailboxv1.FetchMailboxInboxesResponse, error) {
	resp := &mailboxv1.FetchMailboxInboxesResponse{
		MailboxCount: int32(len(targets)),
		Results:      []*mailboxv1.FetchMailboxInboxResult{},
	}
	for _, target := range targets {
		select {
		case <-ctx.Done():
			return nil, status.Error(codes.Canceled, "request cancelled")
		default:
		}

		result := &mailboxv1.FetchMailboxInboxResult{Mailbox: publicMailbox(target.resultMailbox)}
		messages, err := s.watcher.FetchMailboxInbox(ctx, target.fetchMailbox, request.GetLimitPerMailbox(), request.GetReceivedAfterUnix())
		if err != nil {
			result.ErrorMessage = safeMailboxError(err)
			if cached, cacheErr := s.store.ListInboxMessagesSince(ctx, target.fetchMailbox.GetEmailAddress(), request.GetLimitPerMailbox(), request.GetReceivedAfterUnix()); cacheErr == nil {
				result.Messages = cached
				resp.MessageCount += int32(len(cached))
			}
			resp.FailedCount++
		} else {
			result.Messages = messages
			resp.FetchedCount++
			resp.MessageCount += int32(len(messages))
		}
		resp.Results = append(resp.Results, result)
	}
	return resp, nil
}
