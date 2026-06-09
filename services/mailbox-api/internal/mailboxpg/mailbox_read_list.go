package mailboxpg

import (
	"context"

	"mailboxapi/internal/mailboxmodel"
)

func (r *Repository) ListMailboxes(ctx context.Context, authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxmodel.ListPage, error) {
	query, err := r.newMailboxListQuery(authStatus, provider, emailAddress, cursorValue, limit)
	if err != nil {
		return mailboxmodel.ListPage{}, err
	}
	stored, err := r.listStoredMailboxes(ctx, query)
	if err != nil {
		return mailboxmodel.ListPage{}, err
	}
	rows := stored
	virtual, err := r.listVirtualMailboxes(ctx, query)
	if err != nil {
		return mailboxmodel.ListPage{}, err
	}
	if len(virtual) > 0 {
		rows = append(rows, virtual...)
	}
	return mailboxPageFromRows(rows, query.Limit), nil
}
