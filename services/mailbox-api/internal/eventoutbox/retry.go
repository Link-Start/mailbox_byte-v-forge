package eventoutbox

import (
	"strings"
	"time"
)

func DefaultRetryDelay(attempt int32) time.Duration {
	switch {
	case attempt <= 1:
		return 5 * time.Second
	case attempt == 2:
		return 15 * time.Second
	case attempt == 3:
		return 30 * time.Second
	case attempt <= 6:
		return time.Minute
	default:
		return 5 * time.Minute
	}
}

func TruncateError(err error) string {
	if err == nil {
		return ""
	}
	message := strings.TrimSpace(err.Error())
	if len(message) <= 1000 {
		return message
	}
	return message[:1000]
}
