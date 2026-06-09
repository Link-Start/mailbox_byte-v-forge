package pagex

func TrimLimit[T any](rows []T, limit int) ([]T, bool) {
	if limit < 0 {
		limit = 0
	}
	if len(rows) <= limit {
		return rows, false
	}
	return rows[:limit], true
}
