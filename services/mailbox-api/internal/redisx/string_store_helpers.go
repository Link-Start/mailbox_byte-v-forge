package redisx

import (
	"fmt"
	"strings"
	"time"
)

func (s *StringStore) redisKeys(keys []string) ([]string, []string) {
	seen := map[string]struct{}{}
	cleanKeys := make([]string, 0, len(keys))
	redisKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		cleanKey, redisKey, ok := s.cleanRedisKey(key)
		if !ok {
			continue
		}
		if _, exists := seen[cleanKey]; exists {
			continue
		}
		seen[cleanKey] = struct{}{}
		cleanKeys = append(cleanKeys, cleanKey)
		redisKeys = append(redisKeys, redisKey)
	}
	return cleanKeys, redisKeys
}

func (s *StringStore) redisKey(key string) (string, bool) {
	_, redisKey, ok := s.cleanRedisKey(key)
	return redisKey, ok
}

func (s *StringStore) cleanRedisKey(key string) (string, string, bool) {
	if s == nil || s.client == nil {
		return "", "", false
	}
	return s.keyspace.CleanKey(key)
}

func (s *StringStore) effectiveTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return s.ttl
	}
	return ttl
}

func redisStringValue(raw any) (string, bool) {
	switch value := raw.(type) {
	case nil:
		return "", false
	case string:
		return value, true
	case []byte:
		return string(value), true
	default:
		return fmt.Sprint(value), true
	}
}

func cleanHashFields(fields []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		out = append(out, field)
	}
	return out
}

func cleanHashValues(values map[string]string) map[string]string {
	out := make(map[string]string, len(values))
	for field, value := range values {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		out[field] = value
	}
	return out
}
