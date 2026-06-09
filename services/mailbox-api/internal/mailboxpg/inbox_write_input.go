package mailboxpg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/inboxapp"
)

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
