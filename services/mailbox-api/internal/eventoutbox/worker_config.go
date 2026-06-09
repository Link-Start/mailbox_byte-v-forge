package eventoutbox

import (
	"context"
	"log"
	"strings"
	"time"
)

const (
	DefaultBatch          = 20
	DefaultInterval       = time.Second
	DefaultActiveInterval = 100 * time.Millisecond
)

type PendingProcessor interface {
	PublishPending(ctx context.Context, batch int) (int, error)
}

type WorkerConfig struct {
	Name           string
	Processor      PendingProcessor
	Batch          int
	Interval       time.Duration
	ActiveInterval time.Duration
	Logf           func(string, ...any)
}

func normalizeWorkerConfig(cfg WorkerConfig) WorkerConfig {
	cfg.Name = strings.TrimSpace(cfg.Name)
	if cfg.Name == "" {
		cfg.Name = "event outbox"
	}
	if cfg.Batch <= 0 {
		cfg.Batch = DefaultBatch
	}
	if cfg.Interval <= 0 {
		cfg.Interval = DefaultInterval
	}
	if cfg.ActiveInterval <= 0 {
		cfg.ActiveInterval = DefaultActiveInterval
	}
	if cfg.Logf == nil {
		cfg.Logf = log.Printf
	}
	return cfg
}

func workerDelay(cfg WorkerConfig, published int) time.Duration {
	if published > 0 {
		return cfg.ActiveInterval
	}
	return cfg.Interval
}
