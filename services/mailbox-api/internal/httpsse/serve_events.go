package httpsse

import (
	"context"
	"errors"
	"time"

	observabilityv1 "mailboxapi/internal/contracts/observabilityv1"
	"mailboxapi/internal/hotstream"
)

func serveHotStreamEvents(ctx context.Context, sse *Writer, sub *hotstream.Subscription, eventName string, controlName string, heartbeat time.Duration) {
	sse.Start()
	sse.Event("", controlName, control(observabilityv1.HotStreamControlKind_HOT_STREAM_CONTROL_KIND_CONNECTED, "connected"))
	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-sub.Events:
			if !ok {
				if errors.Is(sub.Err(), hotstream.ErrSlowConsumer) {
					sse.Event("", controlName, control(observabilityv1.HotStreamControlKind_HOT_STREAM_CONTROL_KIND_RESYNC_REQUIRED, "slow consumer; refetch required"))
				}
				return
			}
			sse.Event(event.GetMetadata().GetId(), eventName, event)
		case <-ticker.C:
			sse.Event("", controlName, control(observabilityv1.HotStreamControlKind_HOT_STREAM_CONTROL_KIND_HEARTBEAT, "heartbeat"))
		}
	}
}
