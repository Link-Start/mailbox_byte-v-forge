package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
	"mailboxapi/internal/redisx"
)

func newOptionalRedisClient(ctx context.Context, rawURL string, role string) (*redis.Client, error) {
	role = strings.TrimSpace(role)
	if role == "" {
		role = "runtime"
	}
	if strings.TrimSpace(rawURL) == "" {
		logInfo("mailbox %s redis is not configured; related cache/coordination features run in degraded standalone mode", role)
		return nil, nil
	}
	client, err := redisx.NewClient(ctx, rawURL)
	if err != nil {
		return nil, fmt.Errorf("initialize mailbox %s redis: %w", role, err)
	}
	return client, nil
}
