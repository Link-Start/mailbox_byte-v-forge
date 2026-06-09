package mailboxmem

import (
	"strings"

	"mailboxapi/internal/inboxapp"
)

func messageMatchesKeyword(row inboxapp.MessageRow, keyword string) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	return strings.Contains(strings.ToLower(row.Subject), keyword) ||
		strings.Contains(strings.ToLower(row.BodyPreview), keyword) ||
		strings.Contains(strings.ToLower(row.BodyText), keyword)
}
