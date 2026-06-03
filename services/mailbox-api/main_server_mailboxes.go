package main

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		mailbox.ProviderKey = defaultMailboxProvider()
	} else {
		mailbox.ProviderKey = normalizeMailboxProviderInput(mailbox.GetProviderKey())
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

func (s *server) ListMailboxDomains(ctx context.Context, req *mailboxv1.ListMailboxDomainsRequest) (*mailboxv1.ListMailboxDomainsResponse, error) {
	return s.providers.ListDomains(req), nil
}

func (s *server) SyncMailboxDomains(ctx context.Context, req *mailboxv1.SyncMailboxDomainsRequest) (*mailboxv1.SyncMailboxDomainsResponse, error) {
	return s.providers.SyncDomains(req), nil
}

func (s *server) ListMailboxProviderCapabilities(ctx context.Context, req *mailboxv1.ListMailboxProviderCapabilitiesRequest) (*mailboxv1.ListMailboxProviderCapabilitiesResponse, error) {
	return s.providers.ListCapabilities(req), nil
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
