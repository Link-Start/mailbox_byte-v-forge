package mailboxpg

import (
	"fmt"
	"strings"

	"mailboxapi/internal/mailboxprovider"
)

type providerStorageBuilder struct {
	table       string
	emailColumn string
	columns     []string
	args        []any
	updates     map[string]string
	err         error
}

const (
	excludedUpdate         = "excluded"
	nonEmptyExcludedUpdate = "non_empty_excluded"
)

func newProviderStorageBuilder(fields mailboxprovider.TokenFields) (*providerStorageBuilder, error) {
	table, err := mailboxprovider.SQLIdentifier(fields.Table)
	if err != nil {
		return nil, err
	}
	emailColumn, err := mailboxprovider.SQLIdentifier(fields.EmailColumn)
	if err != nil {
		return nil, err
	}
	return &providerStorageBuilder{table: table, emailColumn: emailColumn, updates: map[string]string{}}, nil
}

func (b *providerStorageBuilder) add(rawColumn string, value any, updateMode string) string {
	if strings.TrimSpace(rawColumn) == "" {
		return ""
	}
	column, err := mailboxprovider.SQLIdentifier(rawColumn)
	if err != nil {
		b.err = err
		return ""
	}
	b.columns = append(b.columns, column)
	b.args = append(b.args, value)
	switch updateMode {
	case excludedUpdate:
		b.updates[column] = fmt.Sprintf("%s = EXCLUDED.%s", column, column)
	case nonEmptyExcludedUpdate:
		b.updates[column] = fmt.Sprintf("%s = CASE WHEN EXCLUDED.%s <> '' THEN EXCLUDED.%s ELSE %s.%s END", column, column, column, b.table, column)
	}
	return column
}
