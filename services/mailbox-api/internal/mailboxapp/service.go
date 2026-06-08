package mailboxapp

import (
	"context"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

	"mailboxapi/internal/mailboxmodel"
)

type Repository interface {
	MarkEmailAuthStatus(ctx context.Context, email string, authStatus string, lastError string) (*mailboxmodel.Record, error)
	UpsertMailbox(ctx context.Context, mailbox *mailboxmodel.Record) (*mailboxmodel.Record, error)
	ListMailboxes(ctx context.Context, authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxmodel.ListPage, error)
	DeleteMailbox(ctx context.Context, email string) (bool, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) MarkEmailAuthStatus(ctx context.Context, request *mailboxv1.MarkEmailAuthStatusRequest) (*mailboxv1.MarkEmailAuthStatusResponse, error) {
	mailbox, err := s.repo.MarkEmailAuthStatus(ctx, request.GetEmailAddress(), mailboxmodel.AuthStatusValue(request.GetAuthStatus()), request.GetLastError())
	if err != nil {
		return nil, err
	}
	return &mailboxv1.MarkEmailAuthStatusResponse{Mailbox: PublicMailbox(mailbox)}, nil
}

func (s *Service) UpsertMailbox(ctx context.Context, request *mailboxv1.UpsertEmailMailboxRequest) (*mailboxv1.UpsertEmailMailboxResponse, error) {
	mailbox, err := s.repo.UpsertMailbox(ctx, RecordFromCredentialInput(request.GetMailbox()))
	if err != nil {
		return nil, err
	}
	return &mailboxv1.UpsertEmailMailboxResponse{Mailbox: PublicMailbox(mailbox)}, nil
}

func (s *Service) ListMailboxes(ctx context.Context, request *mailboxv1.ListEmailMailboxesRequest) (*mailboxv1.ListEmailMailboxesResponse, error) {
	page, err := s.repo.ListMailboxes(ctx, mailboxmodel.AuthStatusValue(request.GetAuthStatus()), request.GetProviderKey(), request.GetEmailAddress(), request.GetCursor(), request.GetLimit())
	if err != nil {
		return nil, err
	}
	return &mailboxv1.ListEmailMailboxesResponse{Mailboxes: PublicMailboxList(page.Mailboxes), NextCursor: page.NextCursor}, nil
}

func (s *Service) DeleteMailbox(ctx context.Context, request *mailboxv1.DeleteMailboxRequest) (*mailboxv1.DeleteMailboxResponse, error) {
	deleted, err := s.repo.DeleteMailbox(ctx, request.GetEmailAddress())
	if err != nil {
		return nil, err
	}
	return &mailboxv1.DeleteMailboxResponse{Deleted: deleted}, nil
}
