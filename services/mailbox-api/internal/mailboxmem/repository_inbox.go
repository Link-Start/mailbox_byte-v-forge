package mailboxmem

import (
	"context"
	"errors"
	"time"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/inboxapp"
)

func (r *Repository) RecordMessages(ctx context.Context, request inboxapp.RecordMessagesRequest) ([]*mailboxv1.EmailInboxMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	provider := r.providers.NormalizeProviderInput(request.Provider)
	if provider == "" {
		return nil, errors.New("email provider is required")
	}
	if len(request.Messages) == 0 {
		return []*mailboxv1.EmailInboxMessage{}, nil
	}

	r.mu.Lock()
	unseen, retention, err := r.recordInboxMessagesLocked(ctx, provider, request, time.Now().Unix())
	if err != nil {
		r.mu.Unlock()
		return nil, err
	}
	r.pruneInboundLocked(provider, retention)
	r.mu.Unlock()
	return unseen, nil
}
