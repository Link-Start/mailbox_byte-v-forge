package mailboxpg

import (
	"strings"

	"mailboxapi/internal/mailboxprovider"
)

func sqlIdentifier(value string) (string, error) {
	return mailboxprovider.SQLIdentifier(value)
}

func sqlStringLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
