package redisx

import "context"

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
