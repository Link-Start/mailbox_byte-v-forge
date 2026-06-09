package eventoutbox

import (
	"context"

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
		ok, err := publishRow(ctx, publisher, row, updates, options)
		if err != nil {
			return published, err
		}
		if ok {
			published++
		}
	}
	return published, nil
}

func publishRow(ctx context.Context, publisher eventbus.Publisher, row Row, updates Updates, options PublishOptions) (bool, error) {
	message, err := MessageFromEnvelope(row.Envelope)
	now := optionNow(options).Unix()
	if err != nil {
		return false, updates.MarkDiscarded(ctx, row.EventID, TruncateError(err), now)
	}
	publishCtx, cancel := context.WithTimeout(ctx, publishTimeout(options))
	_, err = publisher.Publish(publishCtx, message)
	cancel()
	if err != nil {
		nextAttempt := row.AttemptCount + 1
		nextAttemptAt := optionNow(options).Add(retryDelay(options, nextAttempt)).Unix()
		return false, updates.MarkRetry(ctx, row.EventID, nextAttempt, nextAttemptAt, TruncateError(err), optionNow(options).Unix())
	}
	return true, updates.MarkPublished(ctx, row.EventID, optionNow(options).Unix())
}
