package main

import (
	"context"

	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

	"mailboxapi/pb"
)

type server struct {
	pb.UnimplementedMailboxServiceServer

	emailBackend emailBackend
	operations   *operationStore
	activities   *mailboxActivities
	providers    mailboxProviderRuntimeConfig
	hot          *mailboxHotStream
	work         *mailboxWorkDispatcher
}

type emailBackend interface {
	ListMailboxes(context.Context, *pb.ListEmailMailboxesRequest) (*pb.ListEmailMailboxesResponse, error)
	UpsertMailbox(context.Context, *pb.UpsertEmailMailboxRequest) (*pb.UpsertEmailMailboxResponse, error)
	DeleteMailbox(context.Context, *pb.DeleteMailboxRequest) (*pb.DeleteMailboxResponse, error)
	WaitForEmail(context.Context, *mailboxv1.WaitForMailboxEmailRequest) (*mailboxv1.WaitForMailboxEmailResponse, error)
	ListInbox(context.Context, *mailboxv1.ListMailboxInboxRequest) (*mailboxv1.ListMailboxInboxResponse, error)
	FetchInboxes(context.Context, *mailboxv1.FetchMailboxInboxesRequest) (*mailboxv1.FetchMailboxInboxesResponse, error)
	MarkEmailAuthStatus(context.Context, *pb.MarkEmailAuthStatusRequest) (*pb.MarkEmailAuthStatusResponse, error)
}
