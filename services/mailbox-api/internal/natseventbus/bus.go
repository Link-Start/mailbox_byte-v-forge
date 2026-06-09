package natseventbus

import (
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
