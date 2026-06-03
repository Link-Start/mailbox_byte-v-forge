package mailboxprovider

import (
	"fmt"
	"strings"
)

func SQLIdentifier(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("sql identifier is required")
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return "", fmt.Errorf("invalid sql identifier: %s", value)
	}
	return value, nil
}

func validateTokenFields(fields TokenFields) error {
	values := []string{
		fields.Table,
		fields.EmailColumn,
		fields.PasswordColumn,
		fields.RefreshTokenColumn,
		fields.AccessTokenColumn,
		fields.AuthStatusColumn,
		fields.LastErrorColumn,
		fields.CreatedAtColumn,
		fields.UpdatedAtColumn,
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if _, err := SQLIdentifier(value); err != nil {
			return err
		}
	}
	return nil
}
