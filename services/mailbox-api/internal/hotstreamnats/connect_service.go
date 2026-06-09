package hotstreamnats

import (
	"context"
	"errors"
	"strings"

	"github.com/nats-io/nats.go"
	"mailboxapi/internal/hotstream"
)

func ConnectService(ctx context.Context, cfg ServiceConfig, opts ...nats.Option) (*Bus, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		message := strings.TrimSpace(cfg.RequiredMessage)
		if message == "" {
			message = "hotstream nats url is required"
		}
		return nil, errors.New(message)
	}
	clientName := serviceClientName(cfg)
	return Connect(ctx, Config{
		URL:        cfg.URL,
		ClientName: clientName,
		Subject:    serviceSubject(cfg, clientName),
		BufferSize: cfg.BufferSize,
	}, opts...)
}

func serviceClientName(cfg ServiceConfig) string {
	clientName := strings.TrimSpace(cfg.ClientName)
	service := strings.TrimSpace(cfg.Service)
	if clientName == "" {
		clientName = service
	}
	if clientName == "" {
		clientName = "mailbox"
	}
	return clientName
}

func serviceSubject(cfg ServiceConfig, clientName string) string {
	subject := strings.TrimSpace(cfg.Subject)
	if subject != "" {
		return subject
	}
	subjectService := strings.TrimSpace(cfg.Service)
	if subjectService == "" {
		subjectService = clientName
	}
	return hotstream.ServiceStateSubject(subjectService)
}
