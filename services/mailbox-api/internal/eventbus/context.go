package eventbus

import (
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	commonv1 "mailboxapi/internal/contracts/commonv1"
)

const DefaultEventVersion = "v1"
const DefaultEventSpecVersion = "1.0"

type EventMetadataConfig struct {
	EventID        string
	EventName      string
	EventVersion   string
	OccurredAt     time.Time
	SourceService  string
	Subject        string
	CorrelationID  string
	TraceID        string
	IdempotencyKey string
	DataSchema     string
}

func NewEventMetadata(cfg EventMetadataConfig) *commonv1.EventMetadata {
	eventVersion := strings.TrimSpace(cfg.EventVersion)
	if eventVersion == "" {
		eventVersion = DefaultEventVersion
	}
	occurredAt := cfg.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}
	idempotencyKey := strings.TrimSpace(cfg.IdempotencyKey)
	eventID := strings.TrimSpace(cfg.EventID)
	if idempotencyKey == "" {
		idempotencyKey = eventID
	}
	return &commonv1.EventMetadata{
		Id:              eventID,
		Type:            strings.TrimSpace(cfg.EventName),
		Version:         eventVersion,
		Time:            timestamppb.New(occurredAt),
		Source:          strings.TrimSpace(cfg.SourceService),
		CorrelationId:   strings.TrimSpace(cfg.CorrelationID),
		TraceId:         strings.TrimSpace(cfg.TraceID),
		IdempotencyKey:  idempotencyKey,
		Subject:         strings.TrimSpace(cfg.Subject),
		SpecVersion:     DefaultEventSpecVersion,
		DataContentType: ProtobufContentType,
		DataSchema:      strings.TrimSpace(cfg.DataSchema),
	}
}
