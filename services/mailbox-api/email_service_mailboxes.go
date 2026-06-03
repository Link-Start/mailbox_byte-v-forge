package main

import (
	"context"
	"errors"

	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *EmailService) MarkEmailAuthStatus(ctx context.Context, request *mailboxv1.MarkEmailAuthStatusRequest) (*mailboxv1.MarkEmailAuthStatusResponse, error) {
	mailbox, err := s.store.MarkEmailAuthStatus(ctx, request.GetEmailAddress(), mailboxAuthStatusValue(request.GetAuthStatus()), request.GetLastError())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, safeMailboxError(err))
	}
	return &mailboxv1.MarkEmailAuthStatusResponse{Mailbox: publicMailbox(mailbox)}, nil
}

func (s *EmailService) UpsertMailbox(ctx context.Context, request *mailboxv1.UpsertEmailMailboxRequest) (*mailboxv1.UpsertEmailMailboxResponse, error) {
	mailbox, err := s.store.UpsertMailbox(ctx, mailboxRecordFromCredentialInput(request.GetMailbox()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, safeMailboxError(err))
	}
	return &mailboxv1.UpsertEmailMailboxResponse{Mailbox: publicMailbox(mailbox)}, nil
}

func (s *EmailService) ListMailboxes(ctx context.Context, request *mailboxv1.ListEmailMailboxesRequest) (*mailboxv1.ListEmailMailboxesResponse, error) {
	page, err := s.store.ListMailboxes(ctx, mailboxAuthStatusValue(request.GetAuthStatus()), request.GetProviderKey(), request.GetEmailAddress(), request.GetCursor(), request.GetLimit())
	if err != nil {
		if errors.Is(err, errInvalidMailboxListCursor) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, safeMailboxError(err))
	}
	return &mailboxv1.ListEmailMailboxesResponse{Mailboxes: publicMailboxList(page.Mailboxes), NextCursor: page.NextCursor}, nil
}

func (s *EmailService) DeleteMailbox(ctx context.Context, request *mailboxv1.DeleteMailboxRequest) (*mailboxv1.DeleteMailboxResponse, error) {
	deleted, err := s.store.DeleteMailbox(ctx, request.GetEmailAddress())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, safeMailboxError(err))
	}
	return &mailboxv1.DeleteMailboxResponse{Deleted: deleted}, nil
}
