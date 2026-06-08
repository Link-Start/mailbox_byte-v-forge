package main

import (
	"context"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

	"mailboxapi/pb"
)

type server struct {
	pb.UnimplementedMailboxServiceServer

	emailBackend emailBackend
	operations   operationStore
	activities   *mailboxActivities
	providers    mailboxProviderRuntimeConfig
	hot          *mailboxHotStream
	work         *mailboxWorkDispatcher
}

type emailBackend interface {
	ListMailboxes(context.Context, *mailboxv1.ListEmailMailboxesRequest) (*mailboxv1.ListEmailMailboxesResponse, error)
	UpsertMailbox(context.Context, *mailboxv1.UpsertEmailMailboxRequest) (*mailboxv1.UpsertEmailMailboxResponse, error)
	DeleteMailbox(context.Context, *mailboxv1.DeleteMailboxRequest) (*mailboxv1.DeleteMailboxResponse, error)
	WaitForEmail(context.Context, *mailboxv1.WaitForMailboxEmailRequest) (*mailboxv1.WaitForMailboxEmailResponse, error)
	ListInbox(context.Context, *mailboxv1.ListMailboxInboxRequest) (*mailboxv1.ListMailboxInboxResponse, error)
	GetInboxMessage(context.Context, *mailboxv1.GetMailboxInboxMessageRequest) (*mailboxv1.GetMailboxInboxMessageResponse, error)
	FetchInboxes(context.Context, *mailboxv1.FetchMailboxInboxesRequest) (*mailboxv1.FetchMailboxInboxesResponse, error)
	MarkEmailAuthStatus(context.Context, *mailboxv1.MarkEmailAuthStatusRequest) (*mailboxv1.MarkEmailAuthStatusResponse, error)
}
