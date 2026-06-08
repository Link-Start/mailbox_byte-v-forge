package main

import (
	"fmt"
	"net/url"
	"strings"
)

func stringMapValue(data map[string]any, key string) string {
	if data == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(data[key]))
}

func oauthCodeFromURL(value string) (code string, state string, oauthErr string) {
	parsed, err := url.Parse(value)
	if err != nil {
		return "", "", safeMailboxError(err)
	}
	query := parsed.Query()
	return strings.TrimSpace(query.Get("code")), strings.TrimSpace(query.Get("state")), strings.TrimSpace(query.Get("error_description"))
}

func splitScopes(value string) []string {
	parts := strings.Fields(strings.ReplaceAll(value, ",", " "))
	if len(parts) == 0 {
		return strings.Fields(defaultOutlookOAuthScopes)
	}
	return parts
}

func sanitizeURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil {
		return "<invalid-url>"
	}
	query := parsed.Query()
	for _, key := range []string{"code", "state", "session_state"} {
		if query.Has(key) {
			query.Set(key, "***")
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
