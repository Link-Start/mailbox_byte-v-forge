package eventoutbox

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	commonv1 "mailboxapi/internal/contracts/commonv1"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/eventcatalog"
)

const (
	StatusPending   = "PENDING"
	StatusPublished = "PUBLISHED"
	StatusDiscarded = "DISCARDED"
)

const defaultPublishTimeout = 10 * time.Second

var (
	ErrMissingEventID = errors.New("event outbox event_id is required")
	ErrNilPublisher   = errors.New("event outbox publisher is required")
	ErrNilUpdates     = errors.New("event outbox updates is required")
)

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

type Updates interface {
	MarkPublished(ctx context.Context, eventID string, publishedAt int64) error
	MarkRetry(ctx context.Context, eventID string, attemptCount int32, nextAttemptAt int64, lastError string, updatedAt int64) error
	MarkDiscarded(ctx context.Context, eventID string, lastError string, updatedAt int64) error
}

type PublishOptions struct {
	PublishTimeout time.Duration
	RetryDelay     func(int32) time.Duration
	Now            func() time.Time
}

func NewRecord(message eventbus.Message) (Record, error) {
	envelope, err := eventbus.NewEnvelope(message)
	if err != nil {
		return Record{}, err
	}
	metadata := envelope.GetMetadata()
	if metadata == nil || strings.TrimSpace(metadata.GetId()) == "" {
		return Record{}, ErrMissingEventID
	}
	payload, err := proto.Marshal(envelope)
	if err != nil {
		return Record{}, fmt.Errorf("marshal event outbox envelope: %w", err)
	}
	return Record{
		EventID:        metadata.GetId(),
		Subject:        envelope.GetSubject(),
		EventName:      metadata.GetType(),
		IdempotencyKey: metadata.GetIdempotencyKey(),
		Envelope:       payload,
	}, nil
}

func NewRecordFor(
	definition eventcatalog.Definition,
	event proto.Message,
	metadata *commonv1.EventMetadata,
	attributes map[string]string,
) (Record, error) {
	message, err := definition.NewMessage(event, metadata, attributes)
	if err != nil {
		return Record{}, err
	}
	return NewRecord(message)
}
