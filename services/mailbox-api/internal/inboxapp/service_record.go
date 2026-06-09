package inboxapp

import (
	"context"
	"errors"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

func (s *Service) InboxWatermark(ctx context.Context, email string) (int64, error) {
	return s.repo.InboxWatermark(ctx, email)
}

func (s *Service) HasMessages(ctx context.Context, email string) (bool, error) {
	return s.repo.HasInboxMessages(ctx, email)
}

func (s *Service) RecordMessages(ctx context.Context, provider string, messages []*mailboxv1.EmailInboxMessage, expandRecipients bool) ([]*mailboxv1.EmailInboxMessage, error) {
	return s.RecordMessageInputs(ctx, provider, MessageInputs(messages), expandRecipients)
}

func (s *Service) RecordMessageInputs(ctx context.Context, provider string, messages []MessageInput, expandRecipients bool) ([]*mailboxv1.EmailInboxMessage, error) {
	provider = s.normalizeProvider(provider)
	if provider == "" {
		return nil, errors.New("email provider is required")
	}
	unseen, err := s.repo.RecordMessages(ctx, RecordMessagesRequest{
		Provider:         provider,
		Messages:         messages,
		ExpandRecipients: expandRecipients,
		OutboxTable:      s.outboxTable,
		EventSource:      s.eventSource,
		PrepareUnseen:    s.prepareUnseenMessage,
	})
	if err != nil {
		return nil, err
	}
	s.recordRecentInboxMessages(ctx, unseen)
	return unseen, nil
}
