package main

import (
	"context"
	"strings"
	"time"

	abs "github.com/microsoft/kiota-abstractions-go"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"

	"mailboxapi/internal/inboxapp"
)

var graphMessageSelectFields = []string{"id", "internetMessageId", "subject", "from", "bodyPreview", "body", "toRecipients", "ccRecipients", "bccRecipients", "internetMessageHeaders", "receivedDateTime"}

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
	query := &graphusers.ItemMessagesRequestBuilderGetQueryParameters{
		Top:     &top,
		Orderby: []string{"receivedDateTime desc"},
		Select:  graphMessageSelectFields,
	}
	if filter != "" {
		query.Filter = &filter
	}
	resp, err := client.Me().Messages().Get(ctx, &graphusers.ItemMessagesRequestBuilderGetRequestConfiguration{
		Headers:         graphHTMLBodyHeaders(),
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

func (s *outlookInboxSource) fetchMessageWithGraphSDK(ctx context.Context, accessToken string, messageID string) (graphMessage, bool, error) {
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return graphMessage{}, false, nil
	}
	client, err := newGraphClient(accessToken, s.httpClient)
	if err != nil {
		return graphMessage{}, false, err
	}
	resp, err := client.Me().Messages().ByMessageId(messageID).Get(ctx, &graphusers.ItemMessagesMessageItemRequestBuilderGetRequestConfiguration{
		Headers: graphHTMLBodyHeaders(),
		QueryParameters: &graphusers.ItemMessagesMessageItemRequestBuilderGetQueryParameters{
			Select: graphMessageSelectFields,
		},
	})
	if err != nil {
		return graphMessage{}, false, graphFetchErrorFromSDK(err)
	}
	message, ok := graphMessageFromSDK(resp)
	return message, ok, nil
}

func graphHTMLBodyHeaders() *abs.RequestHeaders {
	headers := abs.NewRequestHeaders()
	headers.Add("Prefer", `outlook.body-content-type="html"`)
	return headers
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
