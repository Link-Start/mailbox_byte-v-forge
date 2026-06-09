package mailboxpg

import (
	"fmt"
	"strings"

	"mailboxapi/internal/emailx"
)

func (q *inboxMessageQuery) WhereMailbox(email string) *inboxMessageQuery {
	q.conditions = append(q.conditions, fmt.Sprintf("mailbox_email = %s", q.addArg(emailx.Normalize(email))))
	return q
}

func (q *inboxMessageQuery) WhereProvider(provider string) *inboxMessageQuery {
	provider = strings.TrimSpace(provider)
	if provider != "" {
		q.conditions = append(q.conditions, fmt.Sprintf("provider = %s", q.addArg(provider)))
	}
	return q
}

func (q *inboxMessageQuery) WhereMessageID(messageID string) *inboxMessageQuery {
	messageID = strings.TrimSpace(messageID)
	if messageID != "" {
		q.conditions = append(q.conditions, fmt.Sprintf("message_id = %s", q.addArg(messageID)))
	}
	return q
}

func (q *inboxMessageQuery) WhereReceivedAfter(receivedAt int64) *inboxMessageQuery {
	if receivedAt > 0 {
		q.conditions = append(q.conditions, fmt.Sprintf("received_at > %s", q.addArg(receivedAt)))
	}
	return q
}

func (q *inboxMessageQuery) WhereReceivedAtOrAfter(receivedAt int64) *inboxMessageQuery {
	if receivedAt > 0 {
		q.conditions = append(q.conditions, fmt.Sprintf("received_at >= %s", q.addArg(receivedAt)))
	}
	return q
}

func (q *inboxMessageQuery) WhereKeyword(keyword string) *inboxMessageQuery {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return q
	}
	placeholder := q.addArg("%" + keyword + "%")
	q.conditions = append(q.conditions, fmt.Sprintf("(subject ILIKE %s OR body_preview ILIKE %s OR body_text ILIKE %s)", placeholder, placeholder, placeholder))
	return q
}
