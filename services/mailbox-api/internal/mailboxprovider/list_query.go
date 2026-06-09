package mailboxprovider

import "mailboxapi/internal/pagex"

type ListQuery struct {
	AuthStatus   string
	Provider     string
	EmailAddress string
	Cursor       pagex.KeysetCursor
	Limit        int
}

func (q ListQuery) ScanLimit() int {
	return pagex.KeysetLookaheadLimit(q.Limit)
}

func (q ListQuery) HasCursor() bool {
	return pagex.HasKeysetCursor(q.Cursor)
}
