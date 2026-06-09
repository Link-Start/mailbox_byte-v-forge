package eventoutbox

import (
	"context"
	"errors"

	"mailboxapi/internal/timex"
)

func RunWorker(ctx context.Context, cfg WorkerConfig) error {
	if cfg.Processor == nil {
		return nil
	}
	cfg = normalizeWorkerConfig(cfg)
	for ctx.Err() == nil {
		published, err := cfg.Processor.PublishPending(ctx, cfg.Batch)
		if err != nil {
			cfg.Logf("publish %s failed: %v", cfg.Name, err)
		}
		if err := timex.Sleep(ctx, workerDelay(cfg, published)); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return err
		}
	}
	return nil
}
