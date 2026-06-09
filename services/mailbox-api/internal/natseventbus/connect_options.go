package natseventbus

import (
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

func defaultConnectOptions(cfg Config) []nats.Option {
	name := strings.TrimSpace(cfg.ClientName)
	if name == "" {
		name = "mailbox"
	}
	return []nats.Option{
		nats.Name(name),
		nats.Timeout(5 * time.Second),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(time.Second),
	}
}
