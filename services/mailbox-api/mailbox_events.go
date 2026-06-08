package main

import (
	"context"
	"fmt"

	"mailboxapi/internal/eventbus"
)

const mailboxEventSource = "mailbox-api"

type mailboxEvents struct {
	publisher eventbus.Publisher
	source    string
}

func newMailboxEvents(publisher eventbus.Publisher) *mailboxEvents {
	if publisher == nil {
		return nil
	}
	return &mailboxEvents{publisher: publisher, source: mailboxEventSource}
}

func (p *mailboxEvents) Publish(ctx context.Context, message eventbus.Message) (eventbus.PublishAck, error) {
	if p == nil || p.publisher == nil {
		return eventbus.PublishAck{}, fmt.Errorf("mailbox event publisher is not configured")
	}
	return p.publisher.Publish(ctx, message)
}
