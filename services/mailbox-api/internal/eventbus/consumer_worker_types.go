package eventbus

import (
	"context"
	"time"
)

type LogFunc func(string, ...any)

type MessageHandler func(context.Context, ReceivedMessage)

type ConsumerWorkerConfig struct {
	Name            string
	Consumer        Consumer
	Handler         MessageHandler
	Batch           int
	FetchErrorDelay time.Duration
	Logf            LogFunc
}
