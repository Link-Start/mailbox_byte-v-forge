package mailboxmem

import (
	"context"
	"errors"
	"sync"

	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/inboxapp"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

type Repository struct {
	mu        sync.RWMutex
	providers *mailboxprovider.Registry
	mailboxes map[string]mailboxEntry
	messages  map[string]storedMessage
}

type mailboxEntry struct {
	record         *mailboxmodel.Record
	inboxWatermark int64
}

type storedMessage struct {
	key       string
	createdAt int64
	updatedAt int64
	row       inboxapp.MessageRow
}

func NewRepository(providers *mailboxprovider.Registry) (*Repository, error) {
	if providers == nil {
		return nil, errors.New("mailbox providers are required")
	}
	return &Repository{
		providers: providers,
		mailboxes: map[string]mailboxEntry{},
		messages:  map[string]storedMessage{},
	}, nil
}

func (r *Repository) Close() {}

func (r *Repository) RunOutboxWorker(context.Context, string, eventbus.Publisher, func(string, ...any)) error {
	return nil
}
