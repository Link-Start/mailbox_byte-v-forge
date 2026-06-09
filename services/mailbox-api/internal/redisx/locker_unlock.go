package redisx

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func (l *Lock) Unlock(ctx context.Context) error {
	if l == nil || l.client == nil || l.key == "" || l.token == "" {
		return nil
	}
	return unlockScript.Run(ctx, l.client, []string{l.key}, l.token).Err()
}

var unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`)
