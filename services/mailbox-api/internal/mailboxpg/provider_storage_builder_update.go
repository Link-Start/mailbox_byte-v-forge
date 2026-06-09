package mailboxpg

import "strings"

func (b *providerStorageBuilder) setUpdate(column string, expression string) {
	if column != "" && strings.TrimSpace(expression) != "" {
		b.updates[column] = expression
	}
}
