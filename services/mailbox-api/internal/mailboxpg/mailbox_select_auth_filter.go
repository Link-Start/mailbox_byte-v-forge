package mailboxpg

import (
	"fmt"
	"strings"

	"mailboxapi/internal/mailboxprovider"
)

func (q *mailboxSelectQuery) providerAuthFilter(provider string, authStatus string) (string, bool, error) {
	definition := q.providers.StorageByKey(provider)
	if definition != nil {
		return q.providerAuthFilterSQL(definition, authStatus)
	}
	parts := []string{}
	for _, definition := range q.providers.StorageExtensions() {
		filter, ok, err := q.providerAuthFilterSQL(definition, authStatus)
		if err != nil {
			return "", false, err
		}
		if ok {
			parts = append(parts, filter)
		}
	}
	if len(parts) == 0 {
		return "", false, nil
	}
	return "(" + strings.Join(parts, " OR ") + ")", true, nil
}

func (q *mailboxSelectQuery) providerAuthFilterSQL(provider mailboxprovider.StorageExtension, authStatus string) (string, bool, error) {
	fields, ok := provider.TokenFields()
	if !ok || strings.TrimSpace(fields.AuthStatusColumn) == "" {
		return "", false, nil
	}
	column, err := sqlIdentifier(fields.AuthStatusColumn)
	if err != nil {
		return "", false, fmt.Errorf("provider %s auth status column: %w", provider.Key(), err)
	}
	return fmt.Sprintf("%s.%s = %s", providerStorageAlias(provider.Key()), column, q.addArg(strings.TrimSpace(authStatus))), true, nil
}
