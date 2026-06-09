package eventbus

import (
	"log"
	"strings"
)

func EventID(message ReceivedMessage) string {
	if message.Envelope == nil || message.Envelope.GetMetadata() == nil {
		return ""
	}
	return message.Envelope.GetMetadata().GetId()
}

func normalizeConsumerWorkerConfig(cfg ConsumerWorkerConfig) ConsumerWorkerConfig {
	cfg.Name = strings.TrimSpace(cfg.Name)
	if cfg.Name == "" {
		cfg.Name = "event consumer"
	}
	if cfg.FetchErrorDelay <= 0 {
		cfg.FetchErrorDelay = DefaultFetchErrorDelay
	}
	cfg.Logf = logger(cfg.Logf)
	return cfg
}

func logger(logf LogFunc) LogFunc {
	if logf != nil {
		return logf
	}
	return log.Printf
}
