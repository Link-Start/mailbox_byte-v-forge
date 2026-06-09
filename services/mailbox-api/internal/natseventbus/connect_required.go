package natseventbus

import (
	"errors"
	"strings"

	"github.com/nats-io/nats.go"
)

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
