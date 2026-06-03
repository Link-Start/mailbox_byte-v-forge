package main

import (
	"context"
	"errors"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"

	"mailboxapi/internal/mailboxmodel"
)

func (w *MailWatcher) fetchMailboxMessages(ctx context.Context, mailbox *mailboxmodel.Record, limit int, receivedAfterNs int64) ([]graphMessage, error) {
	manager := w.oauthManagerForMailbox(mailbox)
	accessToken, err := manager.GetAccessToken(ctx)
	if err != nil {
		w.store.MarkAuthFailed(ctx, mailbox.GetEmailAddress(), err)
		return nil, err
	}
	if err := w.persistTokens(ctx, mailbox, manager); err != nil {
		w.store.MarkAuthFailed(ctx, mailbox.GetEmailAddress(), err)
		return nil, err
	}
	messages, err := w.fetchRecentMessages(ctx, accessToken, limit, receivedAfterNs)
	if err != nil {
		var graphErr *GraphFetchError
		if !errors.As(err, &graphErr) {
			w.store.MarkAuthFailed(ctx, mailbox.GetEmailAddress(), err)
			return nil, err
		}
		if !graphErr.IsAuth() {
			return nil, err
		}
		logInfo("Graph auth error for %s; refreshing token and retrying", emailx.Redact(mailbox.GetEmailAddress()))
		accessToken, err = manager.RefreshAccessToken(ctx)
		if err == nil {
			err = w.persistTokens(ctx, mailbox, manager)
		}
		if err == nil {
			messages, err = w.fetchRecentMessages(ctx, accessToken, limit, receivedAfterNs)
		}
		if err != nil {
			w.store.MarkAuthFailed(ctx, mailbox.GetEmailAddress(), err)
			return nil, err
		}
	}
	return messages, nil
}

func (w *MailWatcher) oauthManagerForMailbox(mailbox *mailboxmodel.Record) *OAuthManager {
	key := emailx.Normalize(mailbox.GetEmailAddress())
	refreshToken := strings.TrimSpace(mailbox.GetRefreshToken())
	w.mu.Lock()
	defer w.mu.Unlock()
	entry, ok := w.oauthManagers[key]
	if !ok || entry.refreshToken != refreshToken {
		entry = oauthEntry{refreshToken: refreshToken, manager: NewOAuthManager(refreshToken)}
		w.oauthManagers[key] = entry
	}
	return entry.manager
}

func (w *MailWatcher) persistTokens(ctx context.Context, mailbox *mailboxmodel.Record, manager *OAuthManager) error {
	refreshToken, accessToken := manager.CurrentTokens()
	if refreshToken != mailbox.GetRefreshToken() || accessToken != mailbox.GetAccessToken() {
		return w.store.UpdateMailboxTokens(ctx, mailbox.GetEmailAddress(), refreshToken, accessToken)
	}
	return nil
}
