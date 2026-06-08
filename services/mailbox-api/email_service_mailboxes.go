package main

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

	"mailboxapi/internal/mailboxmodel"
)

func (s *EmailService) MarkEmailAuthStatus(ctx context.Context, request *mailboxv1.MarkEmailAuthStatusRequest) (*mailboxv1.MarkEmailAuthStatusResponse, error) {
	resp, err := s.mailboxes.MarkEmailAuthStatus(ctx, request)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, safeMailboxError(err))
	}
	return resp, nil
}

func (s *EmailService) UpsertMailbox(ctx context.Context, request *mailboxv1.UpsertEmailMailboxRequest) (*mailboxv1.UpsertEmailMailboxResponse, error) {
	resp, err := s.mailboxes.UpsertMailbox(ctx, request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, safeMailboxError(err))
	}
	return resp, nil
}

func (s *EmailService) ListMailboxes(ctx context.Context, request *mailboxv1.ListEmailMailboxesRequest) (*mailboxv1.ListEmailMailboxesResponse, error) {
	resp, err := s.mailboxes.ListMailboxes(ctx, request)
	if err != nil {
		if errors.Is(err, mailboxmodel.ErrInvalidMailboxListCursor) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, safeMailboxError(err))
	}
	return resp, nil
}

func (s *EmailService) DeleteMailbox(ctx context.Context, request *mailboxv1.DeleteMailboxRequest) (*mailboxv1.DeleteMailboxResponse, error) {
	resp, err := s.mailboxes.DeleteMailbox(ctx, request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, safeMailboxError(err))
	}
	return resp, nil
}
