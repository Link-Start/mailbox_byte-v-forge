package pagex

const (
	DefaultLimit = 100
	MaxLimit     = 500
)

func ClampSize(pageSize int, fallback int, maximum int) int {
	if fallback <= 0 {
		fallback = 50
	}
	if maximum <= 0 {
		maximum = fallback
	}
	if pageSize <= 0 {
		return fallback
	}
	if pageSize > maximum {
		return maximum
	}
	return pageSize
}

func NormalizePageLimit(limit int) int {
	return NormalizeLimit(limit, DefaultLimit, MaxLimit)
}

func NormalizeLimit(limit int, defaultLimit int, maxLimit int) int {
	if defaultLimit <= 0 {
		defaultLimit = DefaultLimit
	}
	if maxLimit <= 0 {
		maxLimit = defaultLimit
	}
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}
