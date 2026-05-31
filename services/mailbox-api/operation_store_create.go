package main

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
)

func (s *operationStore) create(ctx context.Context, operationID, action, emailAddress string) (*mailboxv1.MailboxOperation, error) {
	row := &mailboxOperationRow{
		OperationID:  strings.TrimSpace(operationID),
		Action:       strings.ToUpper(strings.TrimSpace(action)),
		Status:       operationStatusCreated,
		EmailAddress: emailx.Normalize(emailAddress),
		LastStep:     "created",
	}
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, err
	}
	return operationRowToProto(row), nil
}

func (s *operationStore) createRegistration(ctx context.Context, operationID string, importOnly bool) (*mailboxv1.MailboxOperation, error) {
	row := &mailboxOperationRow{
		OperationID: strings.TrimSpace(operationID),
		Action:      operationActionRegisterMailbox,
		Status:      operationStatusCreated,
		LastStep:    "queued",
		ImportOnly:  importOnly,
	}
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, err
	}
	return operationRowToProto(row), nil
}

func (s *operationStore) createOAuth(ctx context.Context, operationID string, emailAddress string, onlyMissing bool, limit int32) (*mailboxv1.MailboxOperation, error) {
	row := &mailboxOperationRow{
		OperationID:  strings.TrimSpace(operationID),
		Action:       operationActionMailboxOAuth,
		Status:       operationStatusCreated,
		EmailAddress: emailx.Normalize(emailAddress),
		LastStep:     "queued",
		OnlyMissing:  onlyMissing,
		Limit:        normalizedLimit(limit),
	}
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, err
	}
	return operationRowToProto(row), nil
}
