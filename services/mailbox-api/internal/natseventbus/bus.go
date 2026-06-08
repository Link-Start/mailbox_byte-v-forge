package natseventbus

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"mailboxapi/internal/eventcatalog"
)

const (
	DefaultURL       = nats.DefaultURL
	DefaultStream    = eventcatalog.StreamName
	DefaultSubject   = eventcatalog.StreamSubject
	DefaultFetchWait = 5 * time.Second
)

type Config struct {
	URL          string
	ClientName   string
	Stream       string
	EnsureStream bool
}

type Bus struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	stream string
}

func Connect(cfg Config, opts ...nats.Option) (*Bus, error) {
	url := strings.TrimSpace(cfg.URL)
	if url == "" {
		url = DefaultURL
	}
	name := strings.TrimSpace(cfg.ClientName)
	if name == "" {
		name = "mailbox"
	}
	options := append([]nats.Option{
		nats.Name(name),
		nats.Timeout(5 * time.Second),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(time.Second),
	}, opts...)
	conn, err := nats.Connect(url, options...)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
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

func ConnectRequired(cfg Config, requiredMessage string, opts ...nats.Option) (*Bus, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		requiredMessage = strings.TrimSpace(requiredMessage)
		if requiredMessage == "" {
			requiredMessage = "nats url is required"
		}
		return nil, errors.New(requiredMessage)
	}
	return Connect(cfg, opts...)
}

func (b *Bus) Close() {
	if b == nil || b.conn == nil {
		return
	}
	b.conn.Drain()
	b.conn.Close()
}

func normalizedStreamName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultStream
	}
	return value
}
