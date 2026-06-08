package main

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

func (s *server) WaitForMailboxEmail(ctx context.Context, req *mailboxv1.WaitForMailboxEmailRequest) (*mailboxv1.WaitForMailboxEmailResponse, error) {
	email := emailx.Normalize(req.GetEmailAddress())
	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "email_address is required")
	}
	resp, err := s.emailBackend.WaitForEmail(ctx, &mailboxv1.WaitForMailboxEmailRequest{
		EmailAddress:    email,
		SubjectKeyword:  strings.TrimSpace(req.GetSubjectKeyword()),
		TimeoutSeconds:  req.GetTimeoutSeconds(),
		IssuedAfterUnix: req.GetIssuedAfterUnix(),
		ParserProfile:   req.GetParserProfile(),
		SignalKind:      req.GetSignalKind(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "wait for mailbox email: %s", safeMailboxError(err))
	}
	if resp == nil {
		return nil, status.Error(codes.Internal, "email service returned empty wait response")
	}
	return resp, nil
}

func (s *server) ListMailboxInbox(ctx context.Context, req *mailboxv1.ListMailboxInboxRequest) (*mailboxv1.ListMailboxInboxResponse, error) {
	resp, err := s.emailBackend.ListInbox(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "list mailbox inbox: %s", safeMailboxError(err))
	}
	if resp == nil || resp.GetResult() == nil {
		return nil, status.Error(codes.Internal, "email service returned empty inbox result")
	}
	return resp, nil
}

func (s *server) GetMailboxInboxMessage(ctx context.Context, req *mailboxv1.GetMailboxInboxMessageRequest) (*mailboxv1.GetMailboxInboxMessageResponse, error) {
	email := emailx.Normalize(req.GetEmailAddress())
	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "email_address is required")
	}
	if strings.TrimSpace(req.GetMessageId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "message_id is required")
	}
	resp, err := s.emailBackend.GetInboxMessage(ctx, &mailboxv1.GetMailboxInboxMessageRequest{
		EmailAddress:  email,
		MessageId:     strings.TrimSpace(req.GetMessageId()),
		ProviderKey:   s.providers.normalizeProviderInput(req.GetProviderKey()),
		ParserProfile: strings.TrimSpace(req.GetParserProfile()),
	})
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "get mailbox inbox message: %s", safeMailboxError(err))
	}
	if resp == nil || resp.GetMessage() == nil {
		return nil, status.Error(codes.Internal, "email service returned empty inbox message")
	}
	return resp, nil
}
