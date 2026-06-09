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

func (r *Repository) persistInboxInput(ctx context.Context, tx pgx.Tx, provider string, request inboxapp.RecordMessagesRequest, input inboxapp.MessageInput, now int64, write *inboxWriteResult) error {
	for _, mailboxEmail := range persistTargetMailboxes(input, request.ExpandRecipients) {
		trackInboxRetention(write, mailboxEmail)
		persisted, key, err := persistInboxMessage(ctx, tx, provider, mailboxEmail, input, now)
		if err != nil {
			return err
		}
		trackInboxWatermark(write.watermarks, mailboxEmail, persisted.GetReceivedAtUnix())
		inserted, err := MarkInboxMessageSeen(ctx, tx, provider, mailboxEmail, key, now)
		if err != nil {
			return err
		}
		if inserted {
			if request.PrepareUnseen != nil {
				if err := request.PrepareUnseen(ctx, persisted); err != nil {
					return err
				}
			}
			write.unseen = append(write.unseen, persisted)
		}
	}
	return nil
}
