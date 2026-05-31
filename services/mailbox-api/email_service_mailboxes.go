package main

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mailboxapi/pb"
)

func (s *EmailService) MarkEmailAuthStatus(ctx context.Context, request *pb.MarkEmailAuthStatusRequest) (*pb.MarkEmailAuthStatusResponse, error) {
	mailbox, err := s.store.MarkEmailAuthStatus(ctx, request.GetEmailAddress(), request.GetAuthStatus(), request.GetLastError())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, safeMailboxError(err))
	}
	return &pb.MarkEmailAuthStatusResponse{Mailbox: mailbox}, nil
}

func (s *EmailService) UpsertMailbox(ctx context.Context, request *pb.UpsertEmailMailboxRequest) (*pb.UpsertEmailMailboxResponse, error) {
	mailbox, err := s.store.UpsertMailbox(ctx, request.GetMailbox())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, safeMailboxError(err))
	}
	return &pb.UpsertEmailMailboxResponse{Mailbox: mailbox}, nil
}

func (s *EmailService) ListMailboxes(ctx context.Context, request *pb.ListEmailMailboxesRequest) (*pb.ListEmailMailboxesResponse, error) {
	page, err := s.store.ListMailboxes(ctx, request.GetAuthStatus(), request.GetProviderKey(), request.GetEmailAddress(), request.GetCursor(), request.GetLimit())
	if err != nil {
		if errors.Is(err, errInvalidMailboxListCursor) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, safeMailboxError(err))
	}
	return &pb.ListEmailMailboxesResponse{Mailboxes: page.Mailboxes, NextCursor: page.NextCursor}, nil
}

func (s *EmailService) DeleteMailbox(ctx context.Context, request *pb.DeleteMailboxRequest) (*pb.DeleteMailboxResponse, error) {
	deleted, err := s.store.DeleteMailbox(ctx, request.GetEmailAddress())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, safeMailboxError(err))
	}
	return &pb.DeleteMailboxResponse{Deleted: deleted}, nil
}
