package redisx

import (
	"context"
	"fmt"
	"time"
)

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
