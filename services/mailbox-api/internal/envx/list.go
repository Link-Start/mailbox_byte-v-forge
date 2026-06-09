package envx

import (
	"strings"
	"unicode"
)

func List(name string) []string {
	value := String(name)
	if value == "" {
		return nil
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			items = append(items, item)
		}
	}
	return items
}
