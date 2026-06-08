package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

const pgUniqueViolationCode = "23505"

func (s *pgOperationStore) create(ctx context.Context, operationID, action, emailAddress string) (*mailboxv1.MailboxOperation, error) {
	return s.insert(ctx, mailboxOperationRow{
		OperationID:  strings.TrimSpace(operationID),
		Action:       strings.ToUpper(strings.TrimSpace(action)),
		Status:       operationStatusCreated,
		EmailAddress: emailx.Normalize(emailAddress),
		LastStep:     "created",
	})
}

func (s *pgOperationStore) createRegistration(ctx context.Context, operationID string, importOnly bool) (*mailboxv1.MailboxOperation, error) {
	return s.insert(ctx, mailboxOperationRow{
		OperationID: strings.TrimSpace(operationID),
		Action:      operationActionRegisterMailbox,
		Status:      operationStatusCreated,
		LastStep:    "queued",
		ImportOnly:  importOnly,
	})
}

func (s *pgOperationStore) createOAuth(ctx context.Context, operationID string, emailAddress string, onlyMissing bool, limit int32) (*mailboxv1.MailboxOperation, error) {
	return s.insert(ctx, mailboxOperationRow{
		OperationID:  strings.TrimSpace(operationID),
		Action:       operationActionMailboxOAuth,
		Status:       operationStatusCreated,
		EmailAddress: emailx.Normalize(emailAddress),
		LastStep:     "queued",
		OnlyMissing:  onlyMissing,
		Limit:        normalizedLimit(limit),
	})
}

func (s *pgOperationStore) insert(ctx context.Context, row mailboxOperationRow) (*mailboxv1.MailboxOperation, error) {
	if row.OperationID == "" {
		return nil, errOperationIDRequired
	}
	now := time.Now().Unix()
	row.CreatedAt = now
	row.UpdatedAt = now
	query := `INSERT INTO mailbox_operations (` + operationColumns() + `)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING ` + operationColumns()
	inserted, err := scanOperationRow(s.pool.QueryRow(ctx, query,
		row.OperationID,
		row.Action,
		row.Status,
		row.EmailAddress,
		row.LastStep,
		row.ErrorMessage,
		row.ImportOnly,
		row.OnlyMissing,
		row.Limit,
		row.ClaimOwner,
		row.ClaimUntil,
		row.AttemptCount,
		row.ExitCode,
		row.MailboxCount,
		row.FetchedCount,
		row.FailedCount,
		row.MessageCount,
		row.CreatedAt,
		row.UpdatedAt,
	))
	if err != nil {
		if isPGUniqueViolation(err) {
			return nil, errOperationAlreadyExists
		}
		return nil, err
	}
	return operationRowToProto(&inserted), nil
}

func isPGUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode
}
