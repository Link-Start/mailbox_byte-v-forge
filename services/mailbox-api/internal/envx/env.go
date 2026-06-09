package envx

import (
	"os"
	"strings"
)

func String(name string) string {
	return strings.TrimSpace(os.Getenv(name))
}

func StringDefault(name string, fallback string) string {
	if value := String(name); value != "" {
		return value
	}
	return fallback
}
