package redisx

import "context"

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
