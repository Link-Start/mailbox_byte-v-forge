package mailboxpg

import "fmt"

func (b *providerStorageBuilder) placeholders() []string {
	placeholders := make([]string, 0, len(b.args))
	for index := range b.args {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
	}
	return placeholders
}
