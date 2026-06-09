package secretref

import (
	"strings"

	commonv1 "mailboxapi/internal/contracts/commonv1"
)

func Display(ref *commonv1.SecretRef) string {
	if !Configured(ref) {
		return ""
	}
	provider := strings.TrimSpace(ref.GetProvider())
	purpose := strings.TrimSpace(ref.GetPurpose())
	id := strings.TrimSpace(ref.GetSecretId())
	if len(id) > 12 {
		id = id[:12]
	}
	switch {
	case provider != "" && purpose != "":
		return provider + "/" + purpose + "/" + id
	case provider != "":
		return provider + "/" + id
	default:
		return id
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
