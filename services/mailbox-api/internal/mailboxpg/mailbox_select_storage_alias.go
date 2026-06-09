package mailboxpg

import (
	"strings"

	"mailboxapi/internal/mailboxprovider"
)

func providerStorageAlias(provider string) string {
	provider = mailboxprovider.NormalizeKey(provider)
	var out strings.Builder
	out.WriteString("provider")
	wroteProviderKey := false
	wroteSeparator := false
	for _, r := range provider {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			if !wroteProviderKey {
				out.WriteRune('_')
				wroteProviderKey = true
			}
			out.WriteRune(r)
			wroteSeparator = false
		default:
			if wroteProviderKey && !wroteSeparator {
				out.WriteRune('_')
				wroteSeparator = true
			}
		}
	}
	return strings.TrimRight(out.String(), "_")
}
