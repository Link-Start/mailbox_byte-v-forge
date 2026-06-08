package main

import (
	"errors"
	"net/http"
	"time"

	abs "github.com/microsoft/kiota-abstractions-go"
	"mailboxapi/internal/httpx"
)

func graphFetchErrorFromSDK(err error) error {
	var apiErr abs.ApiErrorable
	if !errors.As(err, &apiErr) {
		return err
	}
	return &GraphFetchError{
		StatusCode: apiErr.GetStatusCode(),
		Body:       safeMailboxError(err),
		RetryAfter: retryAfterFromHeaders(apiErr.GetResponseHeaders()),
	}
}

func retryAfterFromHeaders(headers *abs.ResponseHeaders) time.Duration {
	if headers == nil {
		return 0
	}
	header := http.Header{}
	for _, value := range headers.Get("Retry-After") {
		header.Add("Retry-After", value)
	}
	return httpx.RetryAfterMax(header, 10*time.Second)
}
