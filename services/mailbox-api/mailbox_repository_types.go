package main

import (
	"context"

	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxapp"
	"mailboxapi/internal/mailboxmodel"
)

type mailboxRepository interface {
	mailboxapp.Repository
	inboxapp.Repository
	FindMailbox(ctx context.Context, email string) (*mailboxmodel.Record, error)
	PollMailboxForEmail(ctx context.Context, email string) (*mailboxmodel.Record, error)
	ListOAuthMailboxes(ctx context.Context, limit int32) ([]*mailboxmodel.Record, error)
	UpdateMailboxTokens(ctx context.Context, email string, refreshToken string, accessToken string) error
	RunOutboxWorker(ctx context.Context, table string, publisher eventbus.Publisher, logf func(string, ...any)) error
	Close()
}
