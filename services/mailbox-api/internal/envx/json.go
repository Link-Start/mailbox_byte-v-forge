package envx

import (
	"encoding/json"
	"fmt"
	"strings"
)

func JSONStringMap(name string) (map[string]string, error) {
	value := String(name)
	if value == "" {
		return nil, nil
	}
	items := map[string]string{}
	if err := json.Unmarshal([]byte(value), &items); err != nil {
		return nil, fmt.Errorf("%s must be a JSON object with string values: %w", name, err)
	}
	normalized := make(map[string]string, len(items))
	for key, item := range items {
		key = strings.TrimSpace(key)
		item = strings.TrimSpace(item)
		if key == "" {
			return nil, fmt.Errorf("%s contains an empty key", name)
		}
		if item == "" {
			return nil, fmt.Errorf("%s contains an empty value for key %q", name, key)
		}
		normalized[key] = item
	}
	return normalized, nil
}
