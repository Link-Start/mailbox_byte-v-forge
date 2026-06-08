package redisx

import (
	"context"
	"fmt"
	"time"
)

func (s *StringStore) HashLoadMany(ctx context.Context, key string, fields ...string) (map[string]string, error) {
	redisKey, ok := s.redisKey(key)
	if !ok {
		return map[string]string{}, nil
	}
	cleanFields := cleanHashFields(fields)
	if len(cleanFields) == 0 {
		return map[string]string{}, nil
	}
	values, err := s.client.HMGet(ctx, redisKey, cleanFields...).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(values))
	for idx, raw := range values {
		value, ok := redisStringValue(raw)
		if !ok {
			continue
		}
		out[cleanFields[idx]] = value
	}
	return out, nil
}

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

func (s *StringStore) HashDelete(ctx context.Context, key string, fields ...string) error {
	redisKey, ok := s.redisKey(key)
	if !ok {
		return nil
	}
	cleanFields := cleanHashFields(fields)
	if len(cleanFields) == 0 {
		return nil
	}
	return s.client.HDel(ctx, redisKey, cleanFields...).Err()
}
