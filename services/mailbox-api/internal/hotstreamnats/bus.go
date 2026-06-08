package hotstreamnats

import (
	"github.com/nats-io/nats.go"
	"mailboxapi/internal/hotstream"
)

type Config struct {
	URL        string
	ClientName string
	Subject    string
	BufferSize int
}

type ServiceConfig struct {
	URL             string
	Service         string
	ClientName      string
	Subject         string
	BufferSize      int
	RequiredMessage string
}

type Bus struct {
	conn    *nats.Conn
	hub     *hotstream.Hub
	subject string
	nodeID  string
	sub     *nats.Subscription
}

func (b *Bus) Close() {
	if b == nil {
		return
	}
	if b.sub != nil {
		_ = b.sub.Unsubscribe()
	}
	if b.conn != nil {
		b.conn.Drain()
		b.conn.Close()
	}
}
