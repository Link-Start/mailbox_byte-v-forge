package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type StringStore struct {
	client   redis.Cmdable
	keyspace Keyspace
	ttl      time.Duration
}

func NewStringStore(client redis.Cmdable, prefix string, ttl time.Duration) *StringStore {
	return &StringStore{
		client:   client,
		keyspace: NewKeyspace(prefix),
		ttl:      ttl,
	}
}

func (s *StringStore) DefaultTTL() time.Duration {
	if s == nil {
		return 0
	}
	return s.ttl
}

func (s *StringStore) Load(ctx context.Context, key string) (string, bool, error) {
	redisKey, ok := s.redisKey(key)
	if !ok {
		return "", false, nil
	}
	value, err := s.client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (s *StringStore) LoadMany(ctx context.Context, keys ...string) (map[string]string, error) {
	cleanKeys, redisKeys := s.redisKeys(keys)
	if len(redisKeys) == 0 {
		return map[string]string{}, nil
	}
	values, err := s.client.MGet(ctx, redisKeys...).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(values))
	for idx, raw := range values {
		value, ok := redisStringValue(raw)
		if !ok {
			continue
		}
		out[cleanKeys[idx]] = value
	}
	return out, nil
}

func (s *StringStore) Save(ctx context.Context, key string, value string) error {
	return s.SaveTTL(ctx, key, value, s.ttl)
}

func (s *StringStore) SaveTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	redisKey, ok := s.redisKey(key)
	if !ok {
		return fmt.Errorf("redis string store key is required")
	}
	ttl = s.effectiveTTL(ttl)
	return s.client.Set(ctx, redisKey, value, ttl).Err()
}

func (s *StringStore) Delete(ctx context.Context, key string) error {
	redisKey, ok := s.redisKey(key)
	if !ok {
		return nil
	}
	return s.client.Del(ctx, redisKey).Err()
}
