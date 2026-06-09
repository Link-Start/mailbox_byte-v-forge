package inboxapp

import (
	"context"
	"errors"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

	"mailboxapi/pb"
)

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
	return s.RecordMessageInputs(ctx, provider, inboundEmailInputs(event, provider, recipients, s.inboundReceivedAt(event)), false)
}

func (s *Service) inboundReceivedAt(event *pb.InboundEmailWebhook) int64 {
	receivedAt := event.GetReceivedAtUnix()
	if receivedAt <= 0 {
		receivedAt = s.now().Unix()
	}
	return receivedAt
}
