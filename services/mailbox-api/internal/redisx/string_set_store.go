package redisx

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type StringSetStore struct {
	client   redis.Cmdable
	keyspace Keyspace
}

func NewStringSetStore(client redis.Cmdable, prefix string) *StringSetStore {
	return &StringSetStore{client: client, keyspace: NewKeyspace(prefix)}
}

func (s *StringSetStore) Add(ctx context.Context, key string, members ...string) error {
	redisKey, cleanMembers, ok := s.redisKeyAndMembers(key, members)
	if !ok {
		return nil
	}
	return s.client.SAdd(ctx, redisKey, stringArgs(cleanMembers)...).Err()
}

func (s *StringSetStore) Remove(ctx context.Context, key string, members ...string) error {
	redisKey, cleanMembers, ok := s.redisKeyAndMembers(key, members)
	if !ok {
		return nil
	}
	return s.client.SRem(ctx, redisKey, stringArgs(cleanMembers)...).Err()
}

func (s *StringSetStore) ScanPage(ctx context.Context, key string, cursor uint64, count int64) ([]string, uint64, error) {
	redisKey, ok := s.redisKey(key)
	if !ok {
		return []string{}, 0, nil
	}
	if count <= 0 {
		count = 100
	}
	values, nextCursor, err := s.client.SScan(ctx, redisKey, cursor, "", count).Result()
	if err != nil {
		return nil, 0, err
	}
	return cleanMembers(values), nextCursor, nil
}
