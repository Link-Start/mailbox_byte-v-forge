package mailboxpg

import (
	"strings"

	"mailboxapi/internal/emailx"
)

func DomainForEmail(email string) string {
	_, domain, ok := strings.Cut(emailx.Normalize(email), "@")
	if !ok {
		return ""
	}
	return domain
}
