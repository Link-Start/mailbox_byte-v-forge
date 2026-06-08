package main

import (
	"context"
	"fmt"

	"mailboxapi/internal/eventbus"
)

const mailboxPlatformEventSource = "mailbox-api"

type mailboxPlatformEvents struct {
	publisher eventbus.Publisher
	source    string
}

func newMailboxPlatformEvents(publisher eventbus.Publisher) *mailboxPlatformEvents {
	if publisher == nil {
		return nil
	}
	return &mailboxPlatformEvents{publisher: publisher, source: mailboxPlatformEventSource}
}

func (p *mailboxPlatformEvents) Publish(ctx context.Context, message eventbus.Message) (eventbus.PublishAck, error) {
	if p == nil || p.publisher == nil {
		return eventbus.PublishAck{}, fmt.Errorf("mailbox platform event publisher is not configured")
	}
	return p.publisher.Publish(ctx, message)
}
