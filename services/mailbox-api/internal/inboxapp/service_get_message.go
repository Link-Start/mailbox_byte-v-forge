package inboxapp

import (
	"context"
	"errors"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

func (s *Service) GetMessage(ctx context.Context, email string, messageID string, provider string, parserProfile string) (*mailboxv1.GetMailboxInboxMessageResponse, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	if messageID == "" {
		return nil, errors.New("message_id is required")
	}
	row, ok, err := s.repo.GetInboxRow(ctx, email, messageID, s.normalizeProvider(provider))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("message not found")
	}
	message, err := MessageFromRow(row, parserProfile, s.normalizeProvider)
	if err != nil {
		return nil, err
	}
	if err := s.AttachSignalSecrets(ctx, message); err != nil {
		s.log("get mailbox email signal secret refresh failed email=%s: %v", emailx.Redact(email), err)
	}
	return &mailboxv1.GetMailboxInboxMessageResponse{
		Message:  message,
		BodyText: row.BodyText,
		HtmlBody: row.HTMLBody,
	}, nil
}
