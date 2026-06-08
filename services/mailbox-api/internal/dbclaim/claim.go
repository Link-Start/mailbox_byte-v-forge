package dbclaim

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	StatusColumn       = "status"
	LastStepColumn     = "last_step"
	ErrorMessageColumn = "error_message"
	ClaimOwnerColumn   = "claim_owner"
	ClaimUntilColumn   = "claim_until"
	AttemptCountColumn = "attempt_count"
)

func ForUpdate() clause.Locking {
	return clause.Locking{Strength: "UPDATE"}
}

func NormalizeLeaseSeconds(requested int32, fallback int32, maximum int32) int32 {
	if requested <= 0 {
		requested = fallback
	}
	if maximum > 0 && requested > maximum {
		requested = maximum
	}
	return requested
}

func Until(nowUnix int64, leaseSeconds int32) int64 {
	if leaseSeconds <= 0 {
		return nowUnix
	}
	return nowUnix + int64(leaseSeconds)
}

func ClaimUpdates(status string, lastStep string, errorMessage string, owner string, claimUntil int64) map[string]any {
	return map[string]any{
		StatusColumn:       strings.TrimSpace(status),
		LastStepColumn:     strings.TrimSpace(lastStep),
		ErrorMessageColumn: strings.TrimSpace(errorMessage),
		ClaimOwnerColumn:   strings.TrimSpace(owner),
		ClaimUntilColumn:   claimUntil,
		AttemptCountColumn: gorm.Expr(AttemptCountColumn+" + ?", 1),
	}
}

func ExtendUpdates(claimUntil int64) map[string]any {
	return map[string]any{ClaimUntilColumn: claimUntil}
}

func ClearUpdates() map[string]any {
	return map[string]any{
		ClaimOwnerColumn: "",
		ClaimUntilColumn: int64(0),
	}
}
