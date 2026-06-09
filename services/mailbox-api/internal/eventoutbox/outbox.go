package eventoutbox

import (
	"time"
)

const (
	StatusPending   = "PENDING"
	StatusPublished = "PUBLISHED"
	StatusDiscarded = "DISCARDED"
)

const defaultPublishTimeout = 10 * time.Second

type Record struct {
	EventID        string
	Subject        string
	EventName      string
	IdempotencyKey string
	Envelope       []byte
}

type Row struct {
	EventID      string
	Envelope     []byte
	AttemptCount int32
}

type PublishOptions struct {
	PublishTimeout time.Duration
	RetryDelay     func(int32) time.Duration
	Now            func() time.Time
}
