package mailboxmodel

type ListPage struct {
	Mailboxes  []*Record
	NextCursor string
}
