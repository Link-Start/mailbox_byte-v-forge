package main

import (
	"context"
	"time"

	abs "github.com/microsoft/kiota-abstractions-go"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"

	"mailboxapi/internal/inboxapp"
)

func (s *outlookInboxSource) fetchOnceWithGraphSDK(ctx context.Context, accessToken string, limit int, receivedAfterNs int64) ([]graphMessage, error) {
	client, err := newGraphClient(accessToken, s.httpClient)
	if err != nil {
		return nil, err
	}
	top := int32(inboxapp.MessageLimitValue(int32(limit), s.messageLimit))
	filter := ""
	if receivedAfterNs > 0 {
		filter = "receivedDateTime gt " + time.Unix(0, receivedAfterNs).UTC().Format(time.RFC3339Nano)
	}
	headers := abs.NewRequestHeaders()
	headers.Add("Prefer", `outlook.body-content-type="text"`)
	query := &graphusers.ItemMessagesRequestBuilderGetQueryParameters{
		Top:     &top,
		Orderby: []string{"receivedDateTime desc"},
		Select:  []string{"id", "internetMessageId", "subject", "from", "bodyPreview", "body", "toRecipients", "ccRecipients", "bccRecipients", "internetMessageHeaders", "receivedDateTime"},
	}
	if filter != "" {
		query.Filter = &filter
	}
	resp, err := client.Me().Messages().Get(ctx, &graphusers.ItemMessagesRequestBuilderGetRequestConfiguration{
		Headers:         headers,
		QueryParameters: query,
	})
	if err != nil {
		return nil, graphFetchErrorFromSDK(err)
	}
	if resp == nil {
		return []graphMessage{}, nil
	}
	return graphMessagesFromSDK(resp.GetValue()), nil
}

func stringValueFromPtr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func graphTimeFromSDK(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
