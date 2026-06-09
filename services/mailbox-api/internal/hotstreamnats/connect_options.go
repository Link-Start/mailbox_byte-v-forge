package hotstreamnats

import (
	"time"

	"github.com/nats-io/nats.go"
)

func defaultConnectOptions(name string) []nats.Option {
	return []nats.Option{
		nats.Name(name + " hotstream"),
		nats.Timeout(5 * time.Second),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(time.Second),
	}
}
