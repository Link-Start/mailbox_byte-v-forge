package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/grpcclient"
	"github.com/byte-v-forge/common-lib/hotstream"
	"github.com/byte-v-forge/common-lib/hotstreamnats"
	"github.com/byte-v-forge/common-lib/natseventbus"
	"github.com/byte-v-forge/common-lib/redisx"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func newPlatformEventBus(ctx context.Context, cfg config) (*natseventbus.Bus, func(), error) {
	if strings.TrimSpace(cfg.platformNATSURL) == "" {
		return nil, nil, nil
	}
	bus, err := natseventbus.Connect(natseventbus.Config{
		URL:        cfg.platformNATSURL,
		ClientName: "mailbox-api",
	})
	if err != nil {
		return nil, nil, err
	}
	return bus, bus.Close, nil
}

func newMailboxHotStreamBus(ctx context.Context, cfg config) (hotstream.Bus, func(), error) {
	if strings.TrimSpace(cfg.platformNATSURL) == "" {
		return nil, nil, fmt.Errorf("PLATFORM_NATS_URL is required for mailbox hotstream")
	}
	bus, err := hotstreamnats.Connect(ctx, hotstreamnats.Config{
		URL:        cfg.platformNATSURL,
		ClientName: "mailbox-api",
		Subject:    hotstream.ServiceStateSubject("mailbox"),
	})
	if err != nil {
		return nil, nil, err
	}
	return bus, bus.Close, nil
}

func newRequiredRedisClient(ctx context.Context, redisURL string, requiredMessage string) (*redis.Client, func() error, error) {
	client, err := redisx.NewRequiredClient(ctx, redisURL, requiredMessage)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize redis client: %w", err)
	}
	return client, client.Close, nil
}

func newGRPCClient(name string, addr string) (*grpc.ClientConn, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, fmt.Errorf("%s address is required", name)
	}
	return grpcclient.NewInsecure(addr)
}
