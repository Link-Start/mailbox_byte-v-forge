package redisx

import "context"

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
