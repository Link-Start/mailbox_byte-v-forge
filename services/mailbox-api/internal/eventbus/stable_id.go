package eventbus

import (
	"strings"

	"mailboxapi/internal/hashx"
)

func StableEventID(prefix string, parts ...string) string {
	return strings.TrimSpace(prefix) + hashx.StableParts(parts...)
}
