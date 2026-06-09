package eventbus

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"
	commonv1 "mailboxapi/internal/contracts/commonv1"
)

const ProtobufContentType = "application/x-protobuf"

type Message struct {
	Subject    string
	Event      proto.Message
	Metadata   *commonv1.EventMetadata
	Extensions map[string]string
}

type ReceivedMessage struct {
	Subject    string
	Envelope   *commonv1.EventEnvelope
	Extensions map[string]string
	Attempt    int32
	Ack        func(context.Context) error
	Nak        func(context.Context) error
	NakDelay   func(context.Context, time.Duration) error
	Term       func(context.Context) error
	DeadLetter func(context.Context, string) error
}

type PublishAck struct {
	Stream    string
	Sequence  uint64
	Duplicate bool
}

type Publisher interface {
	Publish(context.Context, Message) (PublishAck, error)
}

type Consumer interface {
	Fetch(context.Context, int) ([]ReceivedMessage, error)
}
