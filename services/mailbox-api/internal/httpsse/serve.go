package httpsse

import (
	"net/http"

	"mailboxapi/internal/hotstream"
)

func ServeHotStream(w http.ResponseWriter, r *http.Request, subscriber hotstream.Subscriber, filter hotstream.Filter, opts ServeOptions) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if subscriber == nil {
		http.Error(w, "hotstream subscriber is not configured", http.StatusServiceUnavailable)
		return
	}
	sse, err := NewWriter(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sub, err := subscriber.Subscribe(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer sub.Close()

	eventName := nonEmpty(opts.EventName, DefaultEventName)
	controlName := nonEmpty(opts.ControlEventName, DefaultControlName)
	heartbeat := opts.Heartbeat
	if heartbeat <= 0 {
		heartbeat = DefaultHeartbeat
	}
	serveHotStreamEvents(r.Context(), sse, sub, eventName, controlName, heartbeat)
}
