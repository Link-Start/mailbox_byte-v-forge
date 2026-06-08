package natseventbus

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

func (b *Bus) EnsureStream(subject string) error {
	if b == nil || b.js == nil {
		return errors.New("nats event bus is not connected")
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		subject = DefaultSubject
	}
	info, err := b.js.StreamInfo(b.stream)
	if errors.Is(err, nats.ErrStreamNotFound) {
		_, addErr := b.js.AddStream(&nats.StreamConfig{
			Name:       b.stream,
			Subjects:   []string{subject},
			Retention:  nats.LimitsPolicy,
			Storage:    nats.FileStorage,
			Duplicates: 10 * time.Minute,
		})
		if addErr != nil {
			return fmt.Errorf("ensure nats stream %s: %w", b.stream, addErr)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect nats stream %s: %w", b.stream, err)
	}
	if streamHasSubject(info.Config.Subjects, subject) {
		return nil
	}
	config := info.Config
	config.Subjects = append(config.Subjects, subject)
	if _, err := b.js.UpdateStream(&config); err != nil {
		return fmt.Errorf("ensure nats stream %s subject %s: %w", b.stream, subject, err)
	}
	return nil
}

func streamHasSubject(subjects []string, subject string) bool {
	for _, existing := range subjects {
		if strings.TrimSpace(existing) == subject {
			return true
		}
	}
	return false
}
