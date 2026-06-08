package natseventbus

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	commonv1 "mailboxapi/internal/contracts/commonv1"
	"mailboxapi/internal/eventbus"
)

func (b *Bus) Publish(ctx context.Context, message eventbus.Message) (eventbus.PublishAck, error) {
	if b == nil || b.js == nil {
		return eventbus.PublishAck{}, errors.New("nats event bus is not connected")
	}
	envelope, err := eventbus.NewEnvelope(message)
	if err != nil {
		return eventbus.PublishAck{}, err
	}
	payload, err := proto.Marshal(envelope)
	if err != nil {
		return eventbus.PublishAck{}, fmt.Errorf("marshal event envelope: %w", err)
	}
	msg := &nats.Msg{
		Subject: envelope.GetSubject(),
		Header:  envelopeHeaders(envelope),
		Data:    payload,
	}
	opts := []nats.PubOpt{nats.Context(ctx)}
	if idempotencyKey := strings.TrimSpace(envelope.GetMetadata().GetIdempotencyKey()); idempotencyKey != "" {
		opts = append(opts, nats.MsgId(idempotencyKey))
	}
	ack, err := b.js.PublishMsg(msg, opts...)
	if err != nil {
		return eventbus.PublishAck{}, fmt.Errorf("publish nats event %s: %w", envelope.GetSubject(), err)
	}
	return eventbus.PublishAck{
		Stream:    ack.Stream,
		Sequence:  ack.Sequence,
		Duplicate: ack.Duplicate,
	}, nil
}

func envelopeHeaders(envelope *commonv1.EventEnvelope) nats.Header {
	headers := nats.Header{}
	if envelope == nil {
		return headers
	}
	headers.Set("Bvf-Event-Subject", envelope.GetSubject())
	headers.Set("Bvf-Event-Type", envelope.GetPayloadType())
	headers.Set("Content-Type", envelope.GetDataContentType())
	if metadata := envelope.GetMetadata(); metadata != nil {
		headers.Set("Bvf-Event-Id", metadata.GetId())
		headers.Set("Bvf-Event-Name", metadata.GetType())
		headers.Set("Bvf-Event-Version", metadata.GetVersion())
		headers.Set("Bvf-Correlation-Id", metadata.GetCorrelationId())
		headers.Set("Bvf-Trace-Id", metadata.GetTraceId())
		headers.Set("Bvf-Idempotency-Key", metadata.GetIdempotencyKey())
	}
	return headers
}
