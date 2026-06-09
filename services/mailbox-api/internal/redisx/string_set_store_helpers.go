package redisx

import "strings"

func (s *StringSetStore) redisKeyAndMembers(key string, members []string) (string, []string, bool) {
	redisKey, ok := s.redisKey(key)
	if !ok {
		return "", nil, false
	}
	clean := cleanMembers(members)
	return redisKey, clean, len(clean) > 0
}

func (s *StringSetStore) redisKey(key string) (string, bool) {
	if s == nil || s.client == nil {
		return "", false
	}
	return s.keyspace.Key(key)
}

func cleanMembers(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func stringArgs(values []string) []interface{} {
	out := make([]interface{}, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}
