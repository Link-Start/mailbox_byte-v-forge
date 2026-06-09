package eventoutbox

import "context"

type Updates interface {
	MarkPublished(ctx context.Context, eventID string, publishedAt int64) error
	MarkRetry(ctx context.Context, eventID string, attemptCount int32, nextAttemptAt int64, lastError string, updatedAt int64) error
	MarkDiscarded(ctx context.Context, eventID string, lastError string, updatedAt int64) error
}
