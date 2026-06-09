package mailboxmem

import (
	"context"
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
	"mailboxapi/internal/pagex"
)

func (r *Repository) ListMailboxes(ctx context.Context, authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxmodel.ListPage, error) {
	if err := ctx.Err(); err != nil {
		return mailboxmodel.ListPage{}, err
	}
	cursor, err := pagex.DecodeKeysetCursor(cursorValue)
	if err != nil {
		return mailboxmodel.ListPage{}, mailboxmodel.ErrInvalidMailboxListCursor
	}
	query := mailboxprovider.ListQuery{
		AuthStatus:   strings.TrimSpace(authStatus),
		Provider:     r.providers.NormalizeProviderInput(provider),
		EmailAddress: emailx.Normalize(emailAddress),
		Cursor:       cursor,
		Limit:        pagex.NormalizePageLimit(int(limit)),
	}
	r.mu.RLock()
	rows := r.listStoredMailboxesLocked(query)
	rows = append(rows, r.listVirtualMailboxesLocked(query)...)
	r.mu.RUnlock()
	return mailboxPageFromRows(rows, query.Limit), nil
}
