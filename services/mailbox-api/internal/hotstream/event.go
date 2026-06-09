package hotstream

import (
	"strings"
	"time"

	observabilityv1 "mailboxapi/internal/contracts/observabilityv1"
)

const SubjectPrefix = "mailbox.hot"
const DataContentType = "application/x-protobuf"

type EventConfig struct {
	EventID       string
	EventType     string
	SourceService string
	ResourceType  string
	ResourceID    string
	Scope         string
	OccurredAt    time.Time
	CorrelationID string
	TraceID       string
	Attributes    map[string]string
}

func NewEvent(cfg EventConfig) *observabilityv1.HotStreamEvent {
	return &observabilityv1.HotStreamEvent{
		Metadata:     eventMetadata(cfg),
		ResourceType: strings.TrimSpace(cfg.ResourceType),
		ResourceId:   strings.TrimSpace(cfg.ResourceID),
		Scope:        strings.TrimSpace(cfg.Scope),
		Attributes:   CleanAttributes(cfg.Attributes),
	}
}

func ServiceStateSubject(service string) string {
	service = strings.Trim(strings.ToLower(strings.TrimSpace(service)), ".")
	if service == "" {
		service = "mailbox"
	}
	return SubjectPrefix + "." + service + ".state"
}
