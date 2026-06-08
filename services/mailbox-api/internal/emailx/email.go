package emailx

import "strings"

func Normalize(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func Domain(email string) string {
	_, domain, ok := strings.Cut(Normalize(email), "@")
	if !ok {
		return ""
	}
	return domain
}

func CanonicalPlusAlias(email string) string {
	normalized := Normalize(email)
	local, domain, ok := strings.Cut(normalized, "@")
	if !ok || local == "" || domain == "" {
		return normalized
	}
	local, _, _ = strings.Cut(local, "+")
	return local + "@" + domain
}

func Redact(email string) string {
	local, domain, ok := strings.Cut(strings.TrimSpace(email), "@")
	if !ok || local == "" {
		return "***"
	}
	if len(local) > 2 {
		return local[:2] + "***@" + domain
	}
	return "***@" + domain
}
