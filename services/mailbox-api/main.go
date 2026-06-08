package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(safeMailboxError(err))
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load mailbox config: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runtime, closeRuntime, err := newMailboxRuntime(ctx, cfg)
	if err != nil {
		return fmt.Errorf("initialize mailbox runtime: %w", err)
	}
	defer closeRuntime()

	group, groupCtx := errgroup.WithContext(ctx)
	if err := startMailboxEventWorkers(
		groupCtx,
		group,
		cfg,
		runtime.eventBus,
		runtime.repository,
		runtime.emailBackend,
		runtime.operations,
		runtime.activities,
	); err != nil {
		return fmt.Errorf("initialize mailbox event workers: %w", err)
	}
	if err := startMailboxServers(groupCtx, group, cfg, runtime.serverRuntime()); err != nil {
		return fmt.Errorf("initialize mailbox servers: %w", err)
	}
	if err := group.Wait(); err != nil {
		stop()
		return err
	}
	return nil
}
