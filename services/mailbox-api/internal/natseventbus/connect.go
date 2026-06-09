package natseventbus

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
)

func Connect(cfg Config, opts ...nats.Option) (*Bus, error) {
	conn, err := connectNATS(cfg, opts...)
	if err != nil {
		return nil, err
	}
	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("initialize jetstream: %w", err)
	}
	bus := &Bus{conn: conn, js: js, stream: normalizedStreamName(cfg.Stream)}
	if cfg.EnsureStream {
		if err := bus.EnsureStream(DefaultSubject); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return bus, nil
}

func connectNATS(cfg Config, opts ...nats.Option) (*nats.Conn, error) {
	url := strings.TrimSpace(cfg.URL)
	if url == "" {
		url = DefaultURL
	}
	options := append(defaultConnectOptions(cfg), opts...)
	conn, err := nats.Connect(url, options...)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}
	return conn, nil
}
