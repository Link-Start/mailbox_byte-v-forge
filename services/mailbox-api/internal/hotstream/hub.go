package hotstream

import (
	"context"
	"errors"
	"sync"

	observabilityv1 "mailboxapi/internal/contracts/observabilityv1"
)

const DefaultBufferSize = 256

var ErrSlowConsumer = errors.New("hotstream slow consumer")

type Publisher interface {
	Publish(context.Context, *observabilityv1.HotStreamEvent) error
}

type Subscriber interface {
	Subscribe(context.Context, Filter) (*Subscription, error)
}

type Bus interface {
	Publisher
	Subscriber
}

type Hub struct {
	mu     sync.Mutex
	subs   map[*subscription]struct{}
	buffer int
}

type Subscription struct {
	Events <-chan *observabilityv1.HotStreamEvent
	hub    *Hub
	inner  *subscription
}

type subscription struct {
	filter Filter
	events chan *observabilityv1.HotStreamEvent
	done   chan struct{}
	err    error
	once   sync.Once
}

func NewHub(buffer int) *Hub {
	if buffer <= 0 {
		buffer = DefaultBufferSize
	}
	return &Hub{subs: map[*subscription]struct{}{}, buffer: buffer}
}
