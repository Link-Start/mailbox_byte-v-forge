package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/byte-v-forge/common-lib/pagex"
	"github.com/byte-v-forge/common-lib/randx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *server) GetMailboxOperation(ctx context.Context, req *mailboxv1.GetMailboxOperationRequest) (*mailboxv1.GetMailboxOperationResponse, error) {
	operationID := strings.TrimSpace(req.GetOperationId())
	if operationID == "" {
		return &mailboxv1.GetMailboxOperationResponse{ErrorMessage: "operation_id is required"}, nil
	}
	operation, err := s.operations.get(ctx, operationID)
	if err != nil {
		return &mailboxv1.GetMailboxOperationResponse{ErrorMessage: safeMailboxError(err)}, nil
	}
	return &mailboxv1.GetMailboxOperationResponse{Operation: operation}, nil
}

func (s *server) ListMailboxOperations(ctx context.Context, req *mailboxv1.ListMailboxOperationsRequest) (*mailboxv1.ListMailboxOperationsResponse, error) {
	operations, err := s.operations.list(ctx, operationListFilter{
		Limit:        int(req.GetLimit()),
		Status:       req.GetStatus(),
		Action:       req.GetAction(),
		EmailAddress: req.GetEmailAddress(),
	})
	if err != nil {
		return &mailboxv1.ListMailboxOperationsResponse{ErrorMessage: safeMailboxError(err)}, nil
	}
	return &mailboxv1.ListMailboxOperationsResponse{Operations: operations}, nil
}

func (s *server) updateOperation(ctx context.Context, operationID string, update operationUpdate) {
	operation, err := s.operations.update(ctx, operationID, update)
	if err != nil {
		log.Printf("update mailbox operation failed operation=%s: %s", operationID, safeMailboxError(err))
		return
	}
	s.hot.PublishOperation(ctx, operation)
}

func normalizedLimit(limit int32) int32 {
	return int32(pagex.NormalizePageLimit(int(limit)))
}

func operationID(prefix string) string {
	if id, err := randx.Hex(8); err == nil {
		return prefix + "-" + id
	}
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
