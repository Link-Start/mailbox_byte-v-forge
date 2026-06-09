package eventoutbox

import "errors"

var (
	ErrMissingEventID = errors.New("event outbox event_id is required")
	ErrNilPublisher   = errors.New("event outbox publisher is required")
	ErrNilUpdates     = errors.New("event outbox updates is required")
)
