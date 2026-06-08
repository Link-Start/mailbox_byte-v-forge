package main

import (
	"context"
	"log"

	"google.golang.org/protobuf/proto"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/eventbus"

	"mailboxapi/pb"
)

func (s *server) runRegistrationLocally(ctx context.Context, operationID string) {
	go func() {
		result := (&mailboxRegistrationWorker{operations: s.operations, activities: s.activities}).handle(
			context.WithoutCancel(ctx),
			&pb.MailboxRegistrationOperationRequest{OperationId: operationID},
			eventbus.ReceivedMessage{},
		)
		logLocalWorkResult("mailbox registration", operationID, result)
	}()
}

func (s *server) runOAuthLocally(ctx context.Context, operationID string) {
	go func() {
		result := (&mailboxOAuthWorker{operations: s.operations, activities: s.activities}).handle(
			context.WithoutCancel(ctx),
			&pb.MailboxOAuthOperationRequest{OperationId: operationID},
			eventbus.ReceivedMessage{},
		)
		logLocalWorkResult("mailbox OAuth", operationID, result)
	}()
}

func (s *server) runInboxFetchLocally(ctx context.Context, operationID string, request *mailboxv1.FetchMailboxInboxesRequest) {
	cloned := &mailboxv1.FetchMailboxInboxesRequest{}
	if request != nil {
		cloned = proto.Clone(request).(*mailboxv1.FetchMailboxInboxesRequest)
	}
	go func() {
		workCtx := context.WithoutCancel(ctx)
		s.publishOperationUpdate(workCtx, operationID, operationUpdate{Status: operationStatusRunning, LastStep: "fetch_inboxes"})
		resp, err := s.emailBackend.FetchInboxes(workCtx, cloned)
		if err != nil {
			s.publishOperationUpdate(workCtx, operationID, operationUpdate{Status: operationStatusFailed, LastStep: "fetch_inboxes", ErrorMessage: safeMailboxError(err)})
			return
		}
		statusValue := operationStatusSucceeded
		if resp.GetFailedCount() > 0 {
			statusValue = operationStatusFailed
		}
		s.publishOperationUpdate(workCtx, operationID, operationUpdate{
			Status:       statusValue,
			LastStep:     "fetch_inboxes",
			MailboxCount: resp.GetMailboxCount(),
			FetchedCount: resp.GetFetchedCount(),
			FailedCount:  resp.GetFailedCount(),
			MessageCount: resp.GetMessageCount(),
		})
	}()
}

func (s *server) publishOperationUpdate(ctx context.Context, operationID string, update operationUpdate) {
	operation, err := s.operations.update(ctx, operationID, update)
	if err != nil {
		log.Printf("update mailbox local operation failed operation=%s: %s", operationID, safeMailboxError(err))
		return
	}
	s.hot.PublishOperation(ctx, operation)
}

func logLocalWorkResult(label string, operationID string, result eventbus.HandlerResult) {
	if result.Action == eventbus.MessageActionNak || result.Action == eventbus.MessageActionTerm {
		log.Printf("local %s operation finished action=%s operation_id=%s label=%s", label, result.Action, operationID, result.Label)
	}
}
