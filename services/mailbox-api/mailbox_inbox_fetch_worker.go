package main

import (
	"context"
	"strings"

	"mailboxapi/internal/eventbus"

	"mailboxapi/pb"
)

type mailboxInboxFetchWorker struct {
	service    *EmailService
	operations *operationStore
}

func runMailboxInboxFetchWorker(ctx context.Context, consumer eventbus.Consumer, service *EmailService, operations *operationStore) error {
	worker := &mailboxInboxFetchWorker{service: service, operations: operations}
	return eventbus.RunTypedConsumerWorker(ctx, eventbus.TypedConsumerWorkerConfig[*pb.MailboxInboxFetchRequest]{
		Name:           "mailbox inbox fetch requests",
		Consumer:       consumer,
		Expected:       mailboxInboxFetchRequested.ExpectedMessage(),
		NewMessage:     func() *pb.MailboxInboxFetchRequest { return &pb.MailboxInboxFetchRequest{} },
		Handler:        worker.handle,
		MalformedLabel: "terminate malformed mailbox inbox fetch request",
	})
}

func (w *mailboxInboxFetchWorker) handle(ctx context.Context, request *pb.MailboxInboxFetchRequest, _ eventbus.ReceivedMessage) eventbus.HandlerResult {
	operationID := strings.TrimSpace(request.GetOperationId())
	if operationID == "" {
		return eventbus.TermResult("terminate mailbox inbox fetch request without operation_id")
	}
	if operation, err := w.operations.get(ctx, operationID); err == nil {
		if operationStatusFinal(operation.GetStatus()) {
			return eventbus.AckResult("ack finalized mailbox inbox fetch request")
		}
	}
	_, _ = w.operations.update(ctx, operationID, operationUpdate{Status: operationStatusRunning, LastStep: "fetch_inboxes"})
	resp, err := w.service.FetchInboxes(ctx, request.GetRequest())
	if err != nil {
		_, _ = w.operations.update(ctx, operationID, operationUpdate{Status: operationStatusFailed, LastStep: "fetch_inboxes", ErrorMessage: safeMailboxError(err)})
		return eventbus.AckResult("ack failed mailbox inbox fetch request")
	}
	statusValue := operationStatusSucceeded
	if resp.GetFailedCount() > 0 {
		statusValue = operationStatusFailed
	}
	_, _ = w.operations.update(ctx, operationID, operationUpdate{
		Status:       statusValue,
		LastStep:     "fetch_inboxes",
		MailboxCount: resp.GetMailboxCount(),
		FetchedCount: resp.GetFetchedCount(),
		FailedCount:  resp.GetFailedCount(),
		MessageCount: resp.GetMessageCount(),
	})
	return eventbus.AckResult("ack mailbox inbox fetch request")
}
