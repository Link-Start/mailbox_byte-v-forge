package main

import (
	"context"

	"github.com/byte-v-forge/common-lib/hotstream"
	"github.com/byte-v-forge/common-lib/hotstreamnats"
	"github.com/byte-v-forge/common-lib/natseventbus"
)

func newPlatformEventBus(_ context.Context, cfg config) (*natseventbus.Bus, func(), error) {
	bus, err := natseventbus.ConnectRequired(natseventbus.Config{
		URL:        cfg.platformNATSURL,
		ClientName: "mailbox-api",
	}, "PLATFORM_NATS_URL is required for mailbox event workers")
	if err != nil {
		return nil, nil, err
	}
	return bus, bus.Close, nil
}

func newMailboxHotStreamBus(ctx context.Context, cfg config) (hotstream.Bus, func(), error) {
	bus, err := hotstreamnats.ConnectService(ctx, hotstreamnats.ServiceConfig{
		URL:             cfg.platformNATSURL,
		ClientName:      "mailbox-api",
		Service:         "mailbox",
		RequiredMessage: "PLATFORM_NATS_URL is required for mailbox hotstream",
	})
	if err != nil {
		return nil, nil, err
	}
	return bus, bus.Close, nil
}
