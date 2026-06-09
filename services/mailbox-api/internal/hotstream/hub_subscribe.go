package hotstream

import (
	"context"

	observabilityv1 "mailboxapi/internal/contracts/observabilityv1"
)

func (h *Hub) Subscribe(ctx context.Context, filter Filter) (*Subscription, error) {
	if h == nil {
		h = NewHub(DefaultBufferSize)
	}
	sub := &subscription{
		filter: filter,
		events: make(chan *observabilityv1.HotStreamEvent, h.buffer),
		done:   make(chan struct{}),
	}
	h.mu.Lock()
	h.subs[sub] = struct{}{}
	h.mu.Unlock()
	go func() {
		<-ctx.Done()
		h.unsubscribe(sub, ctx.Err())
	}()
	return &Subscription{Events: sub.events, hub: h, inner: sub}, nil
}

func (h *Hub) unsubscribe(sub *subscription, err error) {
	if h == nil || sub == nil {
		return
	}
	h.mu.Lock()
	delete(h.subs, sub)
	h.mu.Unlock()
	sub.close(err)
}
