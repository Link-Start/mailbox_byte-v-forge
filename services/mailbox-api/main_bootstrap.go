package main

import (
	"context"
	"strings"

	"mailboxapi/internal/hotstream"
	"mailboxapi/internal/hotstreamnats"
	"mailboxapi/internal/natseventbus"
)

func newPlatformEventBus(_ context.Context, cfg config) (*natseventbus.Bus, func(), error) {
	if strings.TrimSpace(cfg.mailboxNATSURL) == "" {
		logInfo("MAILBOX_NATS_URL is not configured; mailbox platform events and MQ workers are disabled")
		return nil, func() {}, nil
	}
	bus, err := natseventbus.ConnectRequired(natseventbus.Config{
		URL:        cfg.mailboxNATSURL,
		ClientName: "mailbox-api",
	}, "MAILBOX_NATS_URL is required for mailbox event workers")
	if err != nil {
		return nil, nil, err
	}
	return bus, bus.Close, nil
}

func newMailboxHotStreamBus(ctx context.Context, cfg config) (hotstream.Bus, func(), error) {
	if strings.TrimSpace(cfg.mailboxNATSURL) == "" {
		logInfo("MAILBOX_NATS_URL is not configured; mailbox dashboard hotstream uses in-process subscribers")
		return hotstream.NewHub(hotstream.DefaultBufferSize), func() {}, nil
	}
	bus, err := hotstreamnats.ConnectService(ctx, hotstreamnats.ServiceConfig{
		URL:             cfg.mailboxNATSURL,
		ClientName:      "mailbox-api",
		Service:         "mailbox",
		RequiredMessage: "MAILBOX_NATS_URL is required for mailbox hotstream",
	})
	if err != nil {
		return nil, nil, err
	}
	return bus, bus.Close, nil
}
