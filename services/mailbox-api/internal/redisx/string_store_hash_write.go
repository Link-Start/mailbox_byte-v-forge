package redisx

import (
	"context"
	"fmt"
	"time"
)

func (s *StringStore) HashSaveTTL(ctx context.Context, key string, values map[string]string, ttl time.Duration) error {
	redisKey, ok := s.redisKey(key)
	if !ok {
		return fmt.Errorf("redis hash store key is required")
	}
	cleanValues := cleanHashValues(values)
	if len(cleanValues) == 0 {
		return nil
	}
	if err := s.client.HSet(ctx, redisKey, cleanValues).Err(); err != nil {
		return err
	}
	return s.extendTTL(ctx, redisKey, ttl)
}

func (s *StringStore) extendTTL(ctx context.Context, redisKey string, ttl time.Duration) error {
	ttl = s.effectiveTTL(ttl)
	if ttl <= 0 {
		return nil
	}
	current, err := s.client.TTL(ctx, redisKey).Result()
	if err != nil {
		return err
	}
	if current <= 0 || ttl > current {
		return s.client.Expire(ctx, redisKey, ttl).Err()
	}
	return nil
}
