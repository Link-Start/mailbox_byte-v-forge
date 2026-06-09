package redisx

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type TTLFlagStore struct {
	client   redis.Cmdable
	keyspace Keyspace
	ttl      time.Duration
	value    string
}

func NewTTLFlagStore(client redis.Cmdable, prefix string, ttl time.Duration, value string) *TTLFlagStore {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "1"
	}
	return &TTLFlagStore{client: client, keyspace: NewKeyspace(prefix), ttl: ttl, value: value}
}

func (s *TTLFlagStore) Claim(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	redisKey, ok := s.redisKey(key)
	if !ok {
		return false, fmt.Errorf("redis ttl flag key is required")
	}
	return s.client.SetNX(ctx, redisKey, s.value, s.effectiveTTL(ttl)).Result()
}

func (s *TTLFlagStore) Save(ctx context.Context, key string, ttl time.Duration) error {
	redisKey, ok := s.redisKey(key)
	if !ok {
		return fmt.Errorf("redis ttl flag key is required")
	}
	return s.client.Set(ctx, redisKey, s.value, s.effectiveTTL(ttl)).Err()
}

func (s *TTLFlagStore) Delete(ctx context.Context, key string) error {
	redisKey, ok := s.redisKey(key)
	if !ok {
		return nil
	}
	return s.client.Del(ctx, redisKey).Err()
}
