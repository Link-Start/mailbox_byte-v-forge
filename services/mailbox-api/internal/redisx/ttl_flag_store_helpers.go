package redisx

import "time"

func (s *TTLFlagStore) redisKey(key string) (string, bool) {
	if s == nil || s.client == nil {
		return "", false
	}
	return s.keyspace.Key(key)
}

func (s *TTLFlagStore) effectiveTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return s.ttl
	}
	return ttl
}
