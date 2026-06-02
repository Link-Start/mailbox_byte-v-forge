package main

import (
	"context"
	"log"
	"strings"

	"github.com/byte-v-forge/common-lib/eventbus"

	"mailboxapi/pb"
)

type mailboxRegistrationWorker struct {
	operations *operationStore
	activities *mailboxActivities
}

func runMailboxRegistrationWorker(ctx context.Context, consumer eventbus.Consumer, operations *operationStore, activities *mailboxActivities) error {
	worker := &mailboxRegistrationWorker{operations: operations, activities: activities}
	return eventbus.RunTypedConsumerWorker(ctx, eventbus.TypedConsumerWorkerConfig[*pb.MailboxRegistrationOperationRequest]{
		Name:       "mailbox registration requests",
		Consumer:   consumer,
		Expected:   mailboxRegistrationRequested.ExpectedMessage(),
		NewMessage: func() *pb.MailboxRegistrationOperationRequest { return &pb.MailboxRegistrationOperationRequest{} },
		Validate: func(request *pb.MailboxRegistrationOperationRequest) error {
			return validateMailboxOperationID(request.GetOperationId())
		},
		Handler:        worker.handle,
		MalformedLabel: "terminate malformed mailbox registration request",
	})
}

func (w *mailboxRegistrationWorker) handle(ctx context.Context, request *pb.MailboxRegistrationOperationRequest, _ eventbus.ReceivedMessage) eventbus.HandlerResult {
	operationID := strings.TrimSpace(request.GetOperationId())
	start, err := w.operations.startRegistrationWorkerRun(ctx, operationID)
	if err != nil {
		return operationStartErrorResult(operationID, "mailbox registration", err)
	}
	if start.Final {
		return eventbus.AckResult("ack finalized mailbox registration request")
	}
	result, err := w.activities.runMailboxRegistration(ctx, mailboxRegistrationActionInput{OperationID: operationID, ImportOnly: start.ImportOnly})
	if err != nil {
		log.Printf("mailbox registration operation failed operation_id=%s success=%t: %s", operationID, result.Success, safeMailboxError(err))
	}
	return eventbus.AckResult("ack mailbox registration request")
}

type mailboxOAuthWorker struct {
	operations *operationStore
	activities *mailboxActivities
}

func runMailboxOAuthWorker(ctx context.Context, consumer eventbus.Consumer, operations *operationStore, activities *mailboxActivities) error {
	worker := &mailboxOAuthWorker{operations: operations, activities: activities}
	return eventbus.RunTypedConsumerWorker(ctx, eventbus.TypedConsumerWorkerConfig[*pb.MailboxOAuthOperationRequest]{
		Name:       "mailbox OAuth requests",
		Consumer:   consumer,
		Expected:   mailboxOAuthRequested.ExpectedMessage(),
		NewMessage: func() *pb.MailboxOAuthOperationRequest { return &pb.MailboxOAuthOperationRequest{} },
		Validate: func(request *pb.MailboxOAuthOperationRequest) error {
			return validateMailboxOperationID(request.GetOperationId())
		},
		Handler:        worker.handle,
		MalformedLabel: "terminate malformed mailbox OAuth request",
	})
}

func (w *mailboxOAuthWorker) handle(ctx context.Context, request *pb.MailboxOAuthOperationRequest, _ eventbus.ReceivedMessage) eventbus.HandlerResult {
	operationID := strings.TrimSpace(request.GetOperationId())
	start, err := w.operations.startOAuthWorkerRun(ctx, operationID)
	if err != nil {
		return operationStartErrorResult(operationID, "mailbox OAuth", err)
	}
	if start.Final {
		return eventbus.AckResult("ack finalized mailbox OAuth request")
	}
	result := w.activities.runMailboxOAuthAction(ctx, operationID, start.EmailAddress, start.OnlyMissing, normalizedLimit(start.Limit))
	if !result.Success {
		log.Printf("mailbox OAuth operation failed operation_id=%s error=%s", operationID, safeMailboxText(result.ErrorMessage))
	}
	return eventbus.AckResult("ack mailbox OAuth request")
}
