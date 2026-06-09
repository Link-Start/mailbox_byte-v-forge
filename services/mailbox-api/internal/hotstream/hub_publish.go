package hotstream

import (
	"context"

	"google.golang.org/protobuf/proto"
	observabilityv1 "mailboxapi/internal/contracts/observabilityv1"
)

func (h *Hub) Publish(_ context.Context, event *observabilityv1.HotStreamEvent) error {
	if h == nil || event == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for sub := range h.subs {
		if !sub.filter.Match(event) {
			continue
		}
		cloned, _ := proto.Clone(event).(*observabilityv1.HotStreamEvent)
		if cloned == nil {
			continue
		}
		select {
		case sub.events <- cloned:
		default:
			sub.close(ErrSlowConsumer)
			delete(h.subs, sub)
		}
	}
	return nil
}
