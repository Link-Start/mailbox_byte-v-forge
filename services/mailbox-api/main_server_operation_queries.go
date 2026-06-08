package main

import (
	"context"
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

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
