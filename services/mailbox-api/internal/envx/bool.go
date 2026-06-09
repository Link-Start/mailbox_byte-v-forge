package envx

import "strings"

func Bool(name string, fallback bool) bool {
	return ParseBool(String(name), fallback)
}

func ParseBool(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
