package eventoutbox

import (
	"context"
	"time"

	"mailboxapi/internal/eventbus"
)

func PublishRows(ctx context.Context, publisher eventbus.Publisher, rows []Row, updates Updates, options PublishOptions) (int, error) {
	if publisher == nil {
		return 0, ErrNilPublisher
	}
	if updates == nil {
		return 0, ErrNilUpdates
	}
	published := 0
	for _, row := range rows {
		if ctx.Err() != nil {
			return published, ctx.Err()
		}
		message, err := MessageFromEnvelope(row.Envelope)
		now := optionNow(options).Unix()
		if err != nil {
			if updateErr := updates.MarkDiscarded(ctx, row.EventID, TruncateError(err), now); updateErr != nil {
				return published, updateErr
			}
			continue
		}
		publishCtx, cancel := context.WithTimeout(ctx, publishTimeout(options))
		_, err = publisher.Publish(publishCtx, message)
		cancel()
		if err != nil {
			nextAttempt := row.AttemptCount + 1
			nextAttemptAt := optionNow(options).Add(retryDelay(options, nextAttempt)).Unix()
			if updateErr := updates.MarkRetry(ctx, row.EventID, nextAttempt, nextAttemptAt, TruncateError(err), optionNow(options).Unix()); updateErr != nil {
				return published, updateErr
			}
			continue
		}
		if updateErr := updates.MarkPublished(ctx, row.EventID, optionNow(options).Unix()); updateErr != nil {
			return published, updateErr
		}
		published++
	}
	return published, nil
}

func publishTimeout(options PublishOptions) time.Duration {
	if options.PublishTimeout > 0 {
		return options.PublishTimeout
	}
	return defaultPublishTimeout
}

func retryDelay(options PublishOptions, attempt int32) time.Duration {
	if options.RetryDelay != nil {
		return options.RetryDelay(attempt)
	}
	return DefaultRetryDelay(attempt)
}

func optionNow(options PublishOptions) time.Time {
	if options.Now != nil {
		return options.Now()
	}
	return time.Now()
}
