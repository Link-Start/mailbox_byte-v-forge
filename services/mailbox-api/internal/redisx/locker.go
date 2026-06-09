package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"mailboxapi/internal/timex"
)

type BestEffortLocker struct {
	client   redis.Cmdable
	keyspace Keyspace
	ttl      time.Duration
	retry    time.Duration
}

type Lock struct {
	client redis.Cmdable
	key    string
	token  string
}

func NewBestEffortLocker(client redis.Cmdable, prefix string, ttl time.Duration, retry time.Duration) *BestEffortLocker {
	return &BestEffortLocker{client: client, keyspace: NewKeyspace(prefix), ttl: lockTTL(ttl), retry: lockRetry(retry)}
}

func (l *BestEffortLocker) Lock(ctx context.Context, key string) (*Lock, error) {
	redisKey, ok := l.redisKey(key)
	if !ok {
		return nil, fmt.Errorf("redis lock key is required")
	}
	token, err := lockToken()
	if err != nil {
		return nil, err
	}
	for {
		locked, err := l.client.SetNX(ctx, redisKey, token, l.ttl).Result()
		if err != nil {
			return nil, err
		}
		if locked {
			return &Lock{client: l.client, key: redisKey, token: token}, nil
		}
		if err := timex.Sleep(ctx, l.retry); err != nil {
			return nil, err
		}
	}
}
