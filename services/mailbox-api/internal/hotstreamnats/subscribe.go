package hotstreamnats

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	observabilityv1 "mailboxapi/internal/contracts/observabilityv1"
	"mailboxapi/internal/hotstream"
)

func (b *Bus) Subscribe(ctx context.Context, filter hotstream.Filter) (*hotstream.Subscription, error) {
	if b == nil || b.hub == nil {
		return nil, errors.New("hotstream bus is not configured")
	}
	return b.hub.Subscribe(ctx, filter)
}

func (b *Bus) subscribe() error {
	sub, err := b.conn.Subscribe(b.subject, b.receive)
	if err != nil {
		return fmt.Errorf("subscribe hotstream nats subject %s: %w", b.subject, err)
	}
	b.sub = sub
	if err := b.conn.Flush(); err != nil {
		b.Close()
		return fmt.Errorf("flush hotstream nats subscription: %w", err)
	}
	return nil
}

func (b *Bus) receive(msg *nats.Msg) {
	if b == nil || msg == nil || msg.Header.Get("Bvf-Hotstream-Node") == b.nodeID {
		return
	}
	event := &observabilityv1.HotStreamEvent{}
	if err := proto.Unmarshal(msg.Data, event); err != nil {
		return
	}
	_ = b.hub.Publish(context.Background(), event)
}
