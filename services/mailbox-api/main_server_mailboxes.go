package main

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

func (s *server) ListMailboxes(ctx context.Context, req *mailboxv1.ListEmailMailboxesRequest) (*mailboxv1.ListEmailMailboxesResponse, error) {
	resp, err := s.emailBackend.ListMailboxes(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "list mailboxes: %s", safeMailboxError(err))
	}
	if resp == nil {
		return nil, status.Error(codes.Internal, "email service returned empty mailbox list")
	}
	return resp, nil
}

func (s *server) UpsertMailbox(ctx context.Context, req *mailboxv1.UpsertEmailMailboxRequest) (*mailboxv1.UpsertEmailMailboxResponse, error) {
	mailbox := req.GetMailbox()
	if mailbox == nil || emailx.Normalize(mailbox.GetEmailAddress()) == "" {
		return nil, status.Error(codes.InvalidArgument, "mailbox email_address is required")
	}
	mailbox.EmailAddress = emailx.Normalize(mailbox.GetEmailAddress())
	if mailbox.GetProviderKey() == "" {
		mailbox.ProviderKey = s.providers.defaultProvider()
	} else {
		mailbox.ProviderKey = s.providers.normalizeProviderInput(mailbox.GetProviderKey())
	}
	resp, err := s.emailBackend.UpsertMailbox(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "upsert mailbox: %s", safeMailboxError(err))
	}
	if resp == nil || resp.GetMailbox() == nil {
		return nil, status.Error(codes.Internal, "email service returned empty mailbox")
	}
	return resp, nil
}

func (s *server) DeleteMailbox(ctx context.Context, req *mailboxv1.DeleteMailboxRequest) (*mailboxv1.DeleteMailboxResponse, error) {
	email := emailx.Normalize(req.GetEmailAddress())
	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "email_address is required")
	}
	resp, err := s.emailBackend.DeleteMailbox(ctx, &mailboxv1.DeleteMailboxRequest{EmailAddress: email})
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "delete mailbox: %s", safeMailboxError(err))
	}
	if resp == nil {
		return nil, status.Error(codes.Internal, "email service returned empty delete response")
	}
	return resp, nil
}
