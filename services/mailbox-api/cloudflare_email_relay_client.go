package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/httpx"
	"mailboxapi/internal/protojsonx"

	"mailboxapi/pb"
)

const (
	cloudflareRelayTokenHeader = "X-Relay-Token"
	cloudflareRelayMaxBody     = 8 << 20
	cloudflareRelayAttempts    = 2
)

type cloudflareEmailRelayClient struct {
	baseURL    string
	token      string
	timeout    time.Duration
	maxEvents  int
	httpClient *http.Client
}

type cloudflareRelayStatusError struct {
	status int
	body   string
}

func newCloudflareEmailRelayClient(config cloudflareRelayPullConfig) *cloudflareEmailRelayClient {
	if !config.enabled() {
		return nil
	}
	return &cloudflareEmailRelayClient{
		baseURL:    config.baseURL,
		token:      config.token,
		timeout:    config.timeout,
		maxEvents:  config.maxEvents,
		httpClient: &http.Client{Timeout: config.timeout},
	}
}

func (c *cloudflareEmailRelayClient) PullPending(ctx context.Context, recipient string) ([]*pb.InboundEmailWebhook, error) {
	if c == nil {
		return nil, nil
	}
	var response pb.PendingInboundEmailEventsResponse
	query := url.Values{}
	query.Set("limit", fmt.Sprintf("%d", c.maxEvents))
	if email := emailx.Normalize(recipient); email != "" {
		query.Set("recipient", email)
	}
	if err := c.requestWithRetry(ctx, http.MethodGet, "/pending", query, nil, &response); err != nil {
		return nil, err
	}
	if response.GetErrorMessage() != "" {
		return nil, errors.New(response.GetErrorMessage())
	}
	return response.GetEvents(), nil
}

func (c *cloudflareEmailRelayClient) Ack(ctx context.Context, eventIDs []string) error {
	if c == nil || len(eventIDs) == 0 {
		return nil
	}
	var response pb.AckInboundEmailEventsResponse
	request := &pb.AckInboundEmailEventsRequest{EventIds: uniqueRelayEventIDs(eventIDs)}
	if len(request.GetEventIds()) == 0 {
		return nil
	}
	if err := c.requestWithRetry(ctx, http.MethodPost, "/ack", nil, request, &response); err != nil {
		return err
	}
	if response.GetErrorMessage() != "" {
		return errors.New(response.GetErrorMessage())
	}
	return nil
}

func (c *cloudflareEmailRelayClient) requestWithRetry(ctx context.Context, method string, path string, query url.Values, body proto.Message, out proto.Message) error {
	var err error
	for attempt := 1; attempt <= cloudflareRelayAttempts; attempt++ {
		err = c.request(ctx, method, path, query, body, out)
		if err == nil || !retryableCloudflareRelayError(err) || attempt == cloudflareRelayAttempts {
			return err
		}
		if sleepErr := sleepContext(ctx, time.Duration(attempt)*200*time.Millisecond); sleepErr != nil {
			return sleepErr
		}
	}
	return err
}

func (c *cloudflareEmailRelayClient) request(ctx context.Context, method string, path string, query url.Values, body proto.Message, out proto.Message) error {
	requestURL, err := c.requestURL(path, query)
	if err != nil {
		return err
	}
	payload, err := relayPayload(body)
	if err != nil {
		return err
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestCtx, method, requestURL, payload)
	if err != nil {
		return err
	}
	request.Header.Set(cloudflareRelayTokenHeader, c.token)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	raw, err := httpx.ReadLimited(response.Body, cloudflareRelayMaxBody)
	if err != nil {
		return err
	}
	if !httpx.Successful(response.StatusCode) {
		return cloudflareRelayStatusError{status: response.StatusCode, body: safeMailboxText(string(raw))}
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return protojsonx.Unmarshal(raw, out)
}

func (c *cloudflareEmailRelayClient) requestURL(path string, query url.Values) (string, error) {
	base, err := url.Parse(c.baseURL + path)
	if err != nil {
		return "", err
	}
	if query != nil {
		base.RawQuery = query.Encode()
	}
	return base.String(), nil
}

func relayPayload(body proto.Message) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	raw, err := protojsonx.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(raw), nil
}

func retryableCloudflareRelayError(err error) bool {
	var statusError cloudflareRelayStatusError
	if errors.As(err, &statusError) {
		return statusError.status >= 500
	}
	return true
}

func (e cloudflareRelayStatusError) Error() string {
	body := strings.TrimSpace(e.body)
	if body == "" {
		return fmt.Sprintf("cloudflare relay status %d", e.status)
	}
	return fmt.Sprintf("cloudflare relay status %d: %s", e.status, body)
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func uniqueRelayEventIDs(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		eventID := strings.TrimSpace(value)
		if eventID == "" {
			continue
		}
		if _, ok := seen[eventID]; ok {
			continue
		}
		seen[eventID] = struct{}{}
		out = append(out, eventID)
	}
	return out
}
