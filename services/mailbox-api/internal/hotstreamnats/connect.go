package hotstreamnats

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
	"mailboxapi/internal/hotstream"
)

func Connect(ctx context.Context, cfg Config, opts ...nats.Option) (*Bus, error) {
	url := strings.TrimSpace(cfg.URL)
	if url == "" {
		return nil, errors.New("hotstream nats url is required")
	}
	name := strings.TrimSpace(cfg.ClientName)
	if name == "" {
		name = "mailbox-hotstream"
	}
	subject := strings.TrimSpace(cfg.Subject)
	if subject == "" {
		subject = hotstream.ServiceStateSubject(name)
	}
	conn, err := nats.Connect(url, append(defaultConnectOptions(name), opts...)...)
	if err != nil {
		return nil, fmt.Errorf("connect hotstream nats: %w", err)
	}
	bus := &Bus{conn: conn, hub: hotstream.NewHub(cfg.BufferSize), subject: subject, nodeID: nats.NewInbox()}
	if err := bus.subscribe(); err != nil {
		conn.Close()
		return nil, err
	}
	go func() {
		<-ctx.Done()
		bus.Close()
	}()
	return bus, nil
}
