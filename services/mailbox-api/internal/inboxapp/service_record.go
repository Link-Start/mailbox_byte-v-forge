package inboxapp

import (
	"context"
	"errors"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/byte-v-forge/common-lib/stringx"

	"mailboxapi/pb"
)

func (s *Service) InboxWatermark(ctx context.Context, email string) (int64, error) {
	return s.repo.InboxWatermark(ctx, email)
}

func (s *Service) HasMessages(ctx context.Context, email string) (bool, error) {
	return s.repo.HasInboxMessages(ctx, email)
}

func (s *Service) RecordInboundEmail(ctx context.Context, event *pb.InboundEmailWebhook) ([]*mailboxv1.EmailInboxMessage, error) {
	if event == nil {
		return nil, errors.New("email event is required")
	}
	provider := s.normalizeProvider(event.GetProviderKey())
	if provider == "" {
		return nil, errors.New("email event provider is required")
	}
	recipients := UniqueEmails(event.GetRecipients())
	if len(recipients) == 0 {
		return nil, errors.New("email event recipients are required")
	}
	receivedAt := event.GetReceivedAtUnix()
	if receivedAt <= 0 {
		receivedAt = s.now().Unix()
	}
	body := strings.TrimSpace(event.GetTextBody())
	if body == "" {
		body = CompactMessageText(event.GetHtmlBody(), 5000)
	}

	messages := make([]*mailboxv1.EmailInboxMessage, 0, len(recipients))
	for _, recipient := range recipients {
		key := StableMessageKey(provider, recipient, stringx.FirstNonEmpty(event.GetEventId(), event.GetMessageId(), event.GetSubject()))
		messageID := stringx.FirstNonEmpty(event.GetMessageId(), event.GetEventId(), key)
		messages = append(messages, &mailboxv1.EmailInboxMessage{
			Id:                 messageID,
			MailboxEmail:       recipient,
			Subject:            strings.TrimSpace(event.GetSubject()),
			FromAddress:        emailx.Normalize(event.GetFromAddress()),
			BodyPreview:        CompactMessageText(body, 500),
			ReceivedAtUnix:     receivedAt,
			Recipients:         recipients,
			ProviderKey:        provider,
			SourceMailboxEmail: recipient,
			RawSize:            event.GetRawSize(),
		})
	}
	return s.RecordMessages(ctx, provider, messages, false)
}

func (s *Service) RecordMessages(ctx context.Context, provider string, messages []*mailboxv1.EmailInboxMessage, expandRecipients bool) ([]*mailboxv1.EmailInboxMessage, error) {
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
