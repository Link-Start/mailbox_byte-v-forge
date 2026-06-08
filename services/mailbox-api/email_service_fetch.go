package main

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

	"mailboxapi/internal/mailboxapp"
)

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

		result := &mailboxv1.FetchMailboxInboxResult{Mailbox: mailboxapp.PublicMailbox(target.resultMailbox)}
		messages, err := s.watcher.FetchMailboxInbox(ctx, target.fetchMailbox, request.GetLimitPerMailbox(), request.GetReceivedAfterUnix())
		if err != nil {
			result.ErrorMessage = safeMailboxError(err)
			if cached, cacheErr := s.inbox.ListMessagesSince(ctx, target.fetchMailbox.GetEmailAddress(), request.GetLimitPerMailbox(), request.GetReceivedAfterUnix()); cacheErr == nil {
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
