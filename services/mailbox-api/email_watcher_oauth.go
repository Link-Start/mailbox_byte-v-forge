package main

import (
	"context"
	"errors"
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"

	"mailboxapi/internal/mailboxmodel"
)

func (s *outlookInboxSource) FetchInboxMessages(ctx context.Context, mailbox *mailboxmodel.Record, limit int, receivedAfterNs int64) ([]*mailboxv1.EmailInboxMessage, error) {
	messages, err := s.fetchMailboxMessages(ctx, mailbox, limit, receivedAfterNs)
	if err != nil {
		return nil, err
	}
	return inboxMessages(s.ProviderKey(), mailbox.GetEmailAddress(), messages), nil
}

func (s *outlookInboxSource) fetchMailboxMessages(ctx context.Context, mailbox *mailboxmodel.Record, limit int, receivedAfterNs int64) ([]graphMessage, error) {
	manager := s.oauthManagerForMailbox(mailbox)
	accessToken, err := manager.GetAccessToken(ctx)
	if err != nil {
		s.markAuthFailed(ctx, mailbox.GetEmailAddress(), err)
		return nil, err
	}
	if err := s.persistTokens(ctx, mailbox, manager); err != nil {
		s.markAuthFailed(ctx, mailbox.GetEmailAddress(), err)
		return nil, err
	}
	messages, err := s.fetchRecentMessages(ctx, accessToken, limit, receivedAfterNs)
	if err != nil {
		var graphErr *GraphFetchError
		if !errors.As(err, &graphErr) {
			s.markAuthFailed(ctx, mailbox.GetEmailAddress(), err)
			return nil, err
		}
		if !graphErr.IsAuth() {
			return nil, err
		}
		logInfo("Graph auth error for %s; refreshing token and retrying", emailx.Redact(mailbox.GetEmailAddress()))
		accessToken, err = manager.RefreshAccessToken(ctx)
		if err == nil {
			err = s.persistTokens(ctx, mailbox, manager)
		}
		if err == nil {
			messages, err = s.fetchRecentMessages(ctx, accessToken, limit, receivedAfterNs)
		}
		if err != nil {
			s.markAuthFailed(ctx, mailbox.GetEmailAddress(), err)
			return nil, err
		}
	}
	return messages, nil
}

func (s *outlookInboxSource) oauthManagerForMailbox(mailbox *mailboxmodel.Record) *OAuthManager {
	key := emailx.Normalize(mailbox.GetEmailAddress())
	refreshToken := strings.TrimSpace(mailbox.GetRefreshToken())
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.oauthManagers[key]
	if !ok || entry.refreshToken != refreshToken {
		entry = oauthEntry{refreshToken: refreshToken, manager: NewOAuthManager(refreshToken, s.oauthConfig)}
		s.oauthManagers[key] = entry
	}
	return entry.manager
}

func (s *outlookInboxSource) persistTokens(ctx context.Context, mailbox *mailboxmodel.Record, manager *OAuthManager) error {
	refreshToken, accessToken := manager.CurrentTokens()
	if refreshToken != mailbox.GetRefreshToken() || accessToken != mailbox.GetAccessToken() {
		return s.mailboxes.UpdateMailboxTokens(ctx, mailbox.GetEmailAddress(), refreshToken, accessToken)
	}
	return nil
}

func (s *outlookInboxSource) markAuthFailed(ctx context.Context, email string, cause error) {
	if _, err := s.mailboxes.MarkEmailAuthStatus(ctx, email, mailboxmodel.AuthStatusAuthFailed, safeMailboxError(cause)); err != nil {
		logWarning("failed to mark mailbox auth failed for %s: %v", emailx.Redact(email), err)
	}
}
