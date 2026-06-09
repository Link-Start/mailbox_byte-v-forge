package mailboxpg

import (
	"fmt"
	"strings"

	"mailboxapi/internal/mailboxprovider"
)

func coalesceField(current string, next string) string {
	next = strings.TrimSpace(next)
	if next == "" {
		return current
	}
	if current == "''" {
		return fmt.Sprintf("COALESCE(%s, '')", next)
	}
	return fmt.Sprintf("COALESCE(NULLIF(%s, ''), %s)", next, current)
}

func storageJoinSQL(provider mailboxprovider.StorageExtension, fields mailboxprovider.TokenFields, alias string) (string, error) {
	table, err := sqlIdentifier(fields.Table)
	if err != nil {
		return "", fmt.Errorf("provider %s storage table: %w", provider.Key(), err)
	}
	emailColumn, err := sqlIdentifier(fields.EmailColumn)
	if err != nil {
		return "", fmt.Errorf("provider %s email column: %w", provider.Key(), err)
	}
	return fmt.Sprintf(
		"\nLEFT JOIN %s %s ON %s.%s = m.email AND m.provider = %s",
		table,
		alias,
		alias,
		emailColumn,
		sqlStringLiteral(provider.Key()),
	), nil
}

func coalesceStorageField(current string, alias string, column string) (string, error) {
	next, err := storageColumnSQL(alias, column)
	if err != nil {
		return current, err
	}
	return coalesceField(current, next), nil
}

func storageColumnSQL(alias string, column string) (string, error) {
	if strings.TrimSpace(column) == "" {
		return "", nil
	}
	identifier, err := sqlIdentifier(column)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s", alias, identifier), nil
}
