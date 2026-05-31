package main

import (
	"strings"
)

func acceptLanguage(locale string) string {
	normalized := strings.ToLower(strings.TrimSpace(locale))
	if strings.HasPrefix(normalized, "zh") {
		return "zh-CN,zh;q=0.9,en;q=0.8"
	}
	if strings.HasPrefix(normalized, "id") {
		return "id-ID,id;q=0.9,en;q=0.8"
	}
	return "en-US,en;q=0.9"
}
