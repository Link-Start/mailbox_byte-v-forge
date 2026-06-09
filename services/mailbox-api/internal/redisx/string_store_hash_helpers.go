package redisx

import "strings"

func cleanHashFields(fields []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		out = append(out, field)
	}
	return out
}

func cleanHashValues(values map[string]string) map[string]string {
	out := make(map[string]string, len(values))
	for field, value := range values {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		out[field] = value
	}
	return out
}
