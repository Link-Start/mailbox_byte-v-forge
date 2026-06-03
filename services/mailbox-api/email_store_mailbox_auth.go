package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/jackc/pgx/v5"

	"mailboxapi/internal/mailboxmodel"
)

func (s *MailboxStore) MarkEmailAuthStatus(ctx context.Context, email string, authStatus string, lastError string) (*mailboxmodel.Record, error) {
	email = emailx.Normalize(email)
	authStatus = strings.TrimSpace(authStatus)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	if authStatus == "" {
		return nil, errors.New("auth_status is required")
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row, err := scanMailbox(tx.QueryRow(ctx, mailboxSelectSQL()+" WHERE m.email = $1 FOR UPDATE", email))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	lastError = safeMailboxText(lastError)
	if err := mailboxProviderUpdateAuth(ctx, tx, row.Provider, email, authStatus, lastError, now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, "UPDATE mailboxes SET updated_at = $1 WHERE email = $2", now, email); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.FindMailbox(ctx, email)
}

func (s *MailboxStore) FindMailbox(ctx context.Context, email string) (*mailboxmodel.Record, error) {
	row, err := scanMailbox(s.pool.QueryRow(ctx, mailboxSelectSQL()+" WHERE m.email = $1", emailx.Normalize(email)))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	if err != nil {
		return nil, err
	}
	mailbox := row.toRecord()
	return mailbox, nil
}

func (s *MailboxStore) PollMailboxForEmail(ctx context.Context, email string) (*mailboxmodel.Record, error) {
	email = emailx.Normalize(email)
	row, err := scanMailbox(s.pool.QueryRow(ctx, mailboxSelectSQL()+" WHERE m.email = $1", email))
	if errors.Is(err, pgx.ErrNoRows) {
		canonical := emailx.CanonicalPlusAlias(email)
		if canonical != "" && canonical != email {
			row, err = scanMailbox(s.pool.QueryRow(ctx, mailboxSelectSQL()+" WHERE m.email = $1", canonical))
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	if err != nil {
		return nil, err
	}

	if err := mailboxProviderValidatePoll(row); err != nil {
		return nil, err
	}
	return row.toRecord(), nil
}

func (s *MailboxStore) UpdateMailboxTokens(ctx context.Context, email string, refreshToken string, accessToken string) error {
	email = emailx.Normalize(email)
	row, err := scanMailbox(s.pool.QueryRow(ctx, mailboxSelectSQL()+" WHERE m.email = $1", email))
	if err != nil {
		return err
	}
	if err := mailboxProviderUpdateTokens(ctx, s.pool, row.Provider, email, refreshToken, accessToken); err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, "UPDATE mailboxes SET updated_at = $1 WHERE email = $2", time.Now().Unix(), email)
	return err
}

func (s *MailboxStore) MarkAuthFailed(ctx context.Context, email string, err error) {
	if _, updateErr := s.MarkEmailAuthStatus(ctx, email, mailboxmodel.AuthStatusAuthFailed, safeMailboxError(err)); updateErr != nil {
		logWarning("failed to mark mailbox auth failed for %s: %v", emailx.Redact(email), updateErr)
	}
}
