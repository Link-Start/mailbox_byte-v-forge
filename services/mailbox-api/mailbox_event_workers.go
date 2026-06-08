package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"mailboxapi/internal/eventbus"
)

const defaultMailboxWorkRetryDelay = 5 * time.Second

func operationStartErrorResult(operationID string, label string, err error) eventbus.HandlerResult {
	if errors.Is(err, errOperationAlreadyRunning) {
		return eventbus.NakResult(30*time.Second, "delay busy "+label+" request")
	}
	if errors.Is(err, errOperationInvalidAction) || errors.Is(err, errOperationNotFound) {
		log.Printf("%s request is invalid operation_id=%s: %s", label, operationID, safeMailboxError(err))
		return eventbus.TermResult("terminate invalid " + label + " request")
	}
	log.Printf("start %s request failed operation_id=%s: %s", label, operationID, safeMailboxError(err))
	return eventbus.NakResult(defaultMailboxWorkRetryDelay, "retry "+label+" request")
}

func validateMailboxOperationID(operationID string) error {
	if strings.TrimSpace(operationID) == "" {
		return fmt.Errorf("operation_id is required")
	}
	return nil
}

func mailboxPollRetryDelay(err error, configuredInterval int) time.Duration {
	var graphErr *GraphFetchError
	if errors.As(err, &graphErr) && graphErr.RetryAfter > 0 {
		return graphErr.RetryAfter
	}
	return mailboxPollInterval(configuredInterval)
}

func mailboxPollInterval(configuredInterval int) time.Duration {
	if configuredInterval <= 0 {
		return defaultMailboxWorkRetryDelay
	}
	return time.Duration(configuredInterval) * time.Second
}

func deadlineReached(deadlineUnix int64) bool {
	return deadlineUnix > 0 && !time.Now().Before(time.Unix(deadlineUnix, 0))
}
