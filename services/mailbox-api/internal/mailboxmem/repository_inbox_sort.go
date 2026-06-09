package mailboxmem

import "sort"

func sortStoredMessages(rows []storedMessage) {
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].row.ReceivedAtUnix == rows[j].row.ReceivedAtUnix {
			if rows[i].updatedAt == rows[j].updatedAt {
				return rows[i].key > rows[j].key
			}
			return rows[i].updatedAt > rows[j].updatedAt
		}
		return rows[i].row.ReceivedAtUnix > rows[j].row.ReceivedAtUnix
	})
}

func normalizeInboxRowLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	return limit
}
