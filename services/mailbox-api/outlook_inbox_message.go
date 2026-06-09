package main

import (
	"context"
	"errors"
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxmodel"
)

func (s *outlookInboxSource) FetchInboxMessage(ctx context.Context, mailbox *mailboxmodel.Record, messageID string) (inboxapp.MessageInput, bool, error) {
	message, ok, err := s.fetchMailboxMessage(ctx, mailbox, messageID)
	if err != nil || !ok {
		return inboxapp.MessageInput{}, ok, err
	}
	return inboxMessageInput(s.ProviderKey(), mailbox.GetEmailAddress(), message), true, nil
}

func (s *outlookInboxSource) fetchMailboxMessage(ctx context.Context, mailbox *mailboxmodel.Record, messageID string) (graphMessage, bool, error) {
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return graphMessage{}, false, nil
	}
	manager := s.oauthManagerForMailbox(mailbox)
	accessToken, err := manager.GetAccessToken(ctx)
	if err != nil {
		s.markAuthFailed(ctx, mailbox.GetEmailAddress(), err)
		return graphMessage{}, false, err
	}
	if err := s.persistTokens(ctx, mailbox, manager); err != nil {
		s.markAuthFailed(ctx, mailbox.GetEmailAddress(), err)
		return graphMessage{}, false, err
	}
	message, ok, err := s.fetchMessageWithGraphSDK(ctx, accessToken, messageID)
	if err == nil {
		return message, ok, nil
	}
	var graphErr *GraphFetchError
	if !errors.As(err, &graphErr) || !graphErr.IsAuth() {
		return graphMessage{}, false, err
	}
	logInfo("Graph auth error for %s; refreshing token and retrying", emailx.Redact(mailbox.GetEmailAddress()))
	accessToken, err = manager.RefreshAccessToken(ctx)
	if err == nil {
		err = s.persistTokens(ctx, mailbox, manager)
	}
	if err != nil {
		s.markAuthFailed(ctx, mailbox.GetEmailAddress(), err)
		return graphMessage{}, false, err
	}
	message, ok, err = s.fetchMessageWithGraphSDK(ctx, accessToken, messageID)
	if err != nil {
		s.markAuthFailed(ctx, mailbox.GetEmailAddress(), err)
		return graphMessage{}, false, err
	}
	return message, ok, nil
}
