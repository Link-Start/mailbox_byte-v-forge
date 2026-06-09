package httpsse

import (
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"mailboxapi/internal/protojsonx"

	observabilityv1 "mailboxapi/internal/contracts/observabilityv1"
)

func control(kind observabilityv1.HotStreamControlKind, message string) *observabilityv1.HotStreamControlEvent {
	return &observabilityv1.HotStreamControlEvent{Kind: kind, Message: message, OccurredAt: timestamppb.Now()}
}

func protoJSON(message proto.Message) string {
	data, err := protojsonx.Marshal(message)
	if err != nil {
		fallback, _ := protojsonx.Marshal(control(observabilityv1.HotStreamControlKind_HOT_STREAM_CONTROL_KIND_ERROR, err.Error()))
		return string(fallback)
	}
	return string(data)
}

func nonEmpty(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func sanitizeLine(value string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(value)
}
