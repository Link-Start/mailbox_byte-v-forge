package redisx

import "time"

func lockTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return 30 * time.Second
	}
	return ttl
}

func lockRetry(retry time.Duration) time.Duration {
	if retry <= 0 {
		return 100 * time.Millisecond
	}
	return retry
}

func (l *BestEffortLocker) redisKey(key string) (string, bool) {
	if l == nil || l.client == nil {
		return "", false
	}
	return l.keyspace.Key(key)
}
