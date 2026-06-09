package eventbus

import (
	"context"
	"errors"
	"time"

	"mailboxapi/internal/timex"
)

const DefaultFetchErrorDelay = time.Second

func RunConsumerWorker(ctx context.Context, cfg ConsumerWorkerConfig) error {
	if cfg.Consumer == nil || cfg.Handler == nil {
		return nil
	}
	cfg = normalizeConsumerWorkerConfig(cfg)
	for ctx.Err() == nil {
		messages, err := cfg.Consumer.Fetch(ctx, cfg.Batch)
		if err != nil {
			if err := waitAfterFetchError(ctx, cfg, err); err != nil {
				return err
			}
			continue
		}
		for _, message := range messages {
			cfg.Handler(ctx, message)
		}
	}
	return nil
}

func waitAfterFetchError(ctx context.Context, cfg ConsumerWorkerConfig, fetchErr error) error {
	if ctx.Err() != nil {
		return nil
	}
	cfg.Logf("fetch %s failed: %v", cfg.Name, fetchErr)
	if err := timex.Sleep(ctx, cfg.FetchErrorDelay); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		return err
	}
	return nil
}
