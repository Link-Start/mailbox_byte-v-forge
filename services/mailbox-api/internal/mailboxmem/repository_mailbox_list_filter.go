package mailboxmem

import (
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

func matchesMailboxQuery(record *mailboxmodel.Record, query mailboxprovider.ListQuery) bool {
	if record == nil {
		return false
	}
	if query.AuthStatus != "" && strings.TrimSpace(record.GetAuthStatus()) != query.AuthStatus {
		return false
	}
	if query.Provider != "" && mailboxprovider.NormalizeKey(record.GetProviderKey()) != query.Provider {
		return false
	}
	if query.EmailAddress != "" && emailx.Normalize(record.GetEmailAddress()) != query.EmailAddress {
		return false
	}
	if query.HasCursor() {
		updatedAt := record.GetUpdatedAt()
		cursorAt := query.Cursor.UpdatedAt.Unix()
		if updatedAt > cursorAt || (updatedAt == cursorAt && record.GetEmailAddress() >= query.Cursor.ID) {
			return false
		}
	}
	return true
}
