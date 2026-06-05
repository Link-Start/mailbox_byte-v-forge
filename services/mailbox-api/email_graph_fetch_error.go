package main

import (
	"fmt"
	"net/http"
	"time"
)

type GraphFetchError struct {
	StatusCode int
	Body       string
	RetryAfter time.Duration
}

func (e *GraphFetchError) Error() string {
	body := safeMailboxText(e.Body)
	if len(body) > 500 {
		body = body[:500]
	}
	return fmt.Sprintf("status=%d body=%s", e.StatusCode, body)
}

func (e *GraphFetchError) IsAuth() bool {
	return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
}

func (e *GraphFetchError) Retryable() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= http.StatusInternalServerError
}
