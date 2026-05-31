package main

import (
	"context"
	"errors"
	"time"
)

func (w *MailWatcher) fetchRecentMessages(ctx context.Context, accessToken string, limit int, receivedAfterNs int64) ([]graphMessage, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		messages, err := w.fetchOnce(ctx, accessToken, limit, receivedAfterNs)
		if err == nil {
			return messages, nil
		}
		lastErr = err
		var graphErr *GraphFetchError
		if !errors.As(err, &graphErr) || attempt == 2 || !graphErr.Retryable() {
			break
		}
		delay := graphErr.RetryAfter
		if delay <= 0 {
			delay = time.Duration(attempt+1) * 500 * time.Millisecond
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func (w *MailWatcher) fetchOnce(ctx context.Context, accessToken string, limit int, receivedAfterNs int64) ([]graphMessage, error) {
	return w.fetchOnceWithGraphSDK(ctx, accessToken, limit, receivedAfterNs)
}
