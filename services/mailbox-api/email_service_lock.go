package main

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *EmailService) acquireInboxLock(ctx context.Context) (func(), error) {
	if s.inboxLock == nil {
		return nil, status.Error(codes.Unavailable, "mailbox inbox lock is not configured")
	}
	lock, err := s.inboxLock.Lock(ctx, "fetch-inboxes")
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "inbox fetch wait timeout")
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, status.Error(codes.Canceled, "request cancelled")
		}
		return nil, status.Error(codes.Unavailable, safeMailboxError(err))
	}
	return func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := lock.Unlock(unlockCtx); err != nil {
			logWarning("release mailbox inbox lock failed: %v", err)
		}
	}, nil
}
