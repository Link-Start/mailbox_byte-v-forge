package main

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

func (s *server) RegisterMailbox(ctx context.Context, req *mailboxv1.RegisterMailboxRequest) (*mailboxv1.RegisterMailboxResponse, error) {
	operationID := operationID("mailbox-register")
	if _, err := s.operations.createRegistration(ctx, operationID, req.GetImportOnly()); err != nil {
		return nil, status.Errorf(codes.Internal, "create mailbox operation: %s", safeMailboxError(err))
	}
	if s.work == nil {
		s.runRegistrationLocally(ctx, operationID)
		return &mailboxv1.RegisterMailboxResponse{
			OperationId: operationID,
			Started:     true,
		}, nil
	}
	if err := s.work.PublishRegistrationRequested(ctx, operationID); err != nil {
		s.updateOperation(ctx, operationID, operationUpdate{Status: operationStatusFailed, LastStep: "queue_registration", ErrorMessage: safeMailboxError(err)})
		return nil, status.Errorf(codes.Unavailable, "queue mailbox registration: %s", safeMailboxError(err))
	}

	return &mailboxv1.RegisterMailboxResponse{
		OperationId: operationID,
		Started:     true,
	}, nil
}

func (s *server) RunMailboxOAuth(ctx context.Context, req *mailboxv1.StartMailboxOAuthRequest) (*mailboxv1.StartMailboxOAuthResponse, error) {
	operationID := operationID("mailbox-oauth")
	email := emailx.Normalize(req.GetEmailAddress())
	if _, err := s.operations.createOAuth(ctx, operationID, email, req.GetOnlyMissing(), req.GetLimit()); err != nil {
		return nil, status.Errorf(codes.Internal, "create mailbox operation: %s", safeMailboxError(err))
	}
	if s.work == nil {
		s.runOAuthLocally(ctx, operationID)
		return &mailboxv1.StartMailboxOAuthResponse{
			OperationId: operationID,
			Started:     true,
		}, nil
	}
	if err := s.work.PublishOAuthRequested(ctx, operationID); err != nil {
		s.updateOperation(ctx, operationID, operationUpdate{Status: operationStatusFailed, LastStep: "queue_oauth", ErrorMessage: safeMailboxError(err)})
		return nil, status.Errorf(codes.Unavailable, "queue mailbox OAuth: %s", safeMailboxError(err))
	}

	return &mailboxv1.StartMailboxOAuthResponse{
		OperationId: operationID,
		Started:     true,
	}, nil
}

func (s *server) FetchMailboxInboxes(ctx context.Context, req *mailboxv1.FetchMailboxInboxesRequest) (*mailboxv1.FetchMailboxInboxesResponse, error) {
	operationID := operationID("mailbox-inbox")
	email := emailx.Normalize(req.GetEmailAddress())
	if _, err := s.operations.create(ctx, operationID, operationActionFetchInboxes, email); err != nil {
		return nil, status.Errorf(codes.Internal, "create mailbox operation: %s", safeMailboxError(err))
	}
	request := &mailboxv1.FetchMailboxInboxesRequest{
		LimitPerMailbox:   req.GetLimitPerMailbox(),
		MaxMailboxes:      req.GetMaxMailboxes(),
		EmailAddress:      email,
		ParserProfile:     req.GetParserProfile(),
		ReceivedAfterUnix: req.GetReceivedAfterUnix(),
	}
	if s.work == nil {
		s.runInboxFetchLocally(ctx, operationID, request)
		return &mailboxv1.FetchMailboxInboxesResponse{OperationId: operationID}, nil
	}
	if err := s.work.PublishInboxFetchRequested(ctx, operationID, request); err != nil {
		s.updateOperation(ctx, operationID, operationUpdate{Status: operationStatusFailed, LastStep: "queue_fetch_inboxes", ErrorMessage: safeMailboxError(err)})
		return nil, status.Errorf(codes.Unavailable, "queue mailbox inbox fetch: %s", safeMailboxError(err))
	}
	return &mailboxv1.FetchMailboxInboxesResponse{OperationId: operationID}, nil
}
