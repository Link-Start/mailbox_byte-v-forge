package eventoutbox

import (
	"fmt"
	"strings"
)

func postgresIdentifier(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrInvalidTableName
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return "", fmt.Errorf("%w: %s", ErrInvalidTableName, value)
	}
	return value, nil
}
