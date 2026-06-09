package mailboxpg

import (
	"context"

	"github.com/jackc/pgx/v5"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxprovider"
)

type inboxWriteResult struct {
	retention  mailboxprovider.InboxRetention
	watermarks map[string]int64
	unseen     []*mailboxv1.EmailInboxMessage
}

func (r *Repository) persistInboxInputs(ctx context.Context, tx pgx.Tx, provider string, request inboxapp.RecordMessagesRequest, now int64) (inboxWriteResult, error) {
	write := inboxWriteResult{
		retention: mailboxprovider.InboxRetention{
			TouchedMailboxes: map[string]struct{}{},
			TouchedDomains:   map[string]struct{}{},
		},
		watermarks: map[string]int64{},
		unseen:     []*mailboxv1.EmailInboxMessage{},
	}
	for _, input := range request.Messages {
		if input.Message == nil {
			continue
		}
		if err := r.persistInboxInput(ctx, tx, provider, request, input, now, &write); err != nil {
			return inboxWriteResult{}, err
		}
	}
	return write, nil
}
