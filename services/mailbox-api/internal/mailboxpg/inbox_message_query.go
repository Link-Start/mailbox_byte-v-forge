package mailboxpg

type inboxMessageQuery struct {
	conditions []string
	args       []any
	orderBy    string
	limit      int
}

func newInboxMessageQuery() *inboxMessageQuery {
	return &inboxMessageQuery{}
}
