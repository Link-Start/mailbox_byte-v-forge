package mailboxmem

import (
	"context"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxprovider"
)

func (r *Repository) recordInboxMessagesLocked(ctx context.Context, provider string, request inboxapp.RecordMessagesRequest, now int64) ([]*mailboxv1.EmailInboxMessage, mailboxprovider.InboxRetention, error) {
	unseen := []*mailboxv1.EmailInboxMessage{}
	retention := mailboxprovider.InboxRetention{
		TouchedMailboxes: map[string]struct{}{},
		TouchedDomains:   map[string]struct{}{},
	}
	for _, input := range request.Messages {
		if input.Message == nil {
			continue
		}
		for _, mailboxEmail := range persistTargetMailboxes(input, request.ExpandRecipients) {
			trackInboxRetention(&retention, mailboxEmail)
			persisted, key, row, err := r.prepareInboxMessage(provider, mailboxEmail, input, now)
			if err != nil {
				return nil, retention, err
			}
			storageKey := messageStorageKey(provider, mailboxEmail, key)
			_, existed := r.messages[storageKey]
			r.messages[storageKey] = storedMessage{key: key, createdAt: now, updatedAt: now, row: row}
			r.trackInboxWatermarkLocked(mailboxEmail, persisted.GetReceivedAtUnix(), now)
			if !existed {
				if request.PrepareUnseen != nil {
					if err := request.PrepareUnseen(ctx, persisted); err != nil {
						return nil, retention, err
					}
				}
				unseen = append(unseen, persisted)
			}
		}
	}
	return unseen, retention, nil
}
