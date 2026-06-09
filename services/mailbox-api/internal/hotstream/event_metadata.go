package hotstream

import (
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	commonv1 "mailboxapi/internal/contracts/commonv1"
)

func eventMetadata(cfg EventConfig) *commonv1.EventMetadata {
	occurredAt := cfg.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}
	return &commonv1.EventMetadata{
		Id:              strings.TrimSpace(cfg.EventID),
		Type:            strings.TrimSpace(cfg.EventType),
		Version:         "v1",
		Time:            timestamppb.New(occurredAt),
		Source:          strings.TrimSpace(cfg.SourceService),
		CorrelationId:   strings.TrimSpace(cfg.CorrelationID),
		TraceId:         strings.TrimSpace(cfg.TraceID),
		SpecVersion:     "1.0",
		DataContentType: DataContentType,
	}
}
