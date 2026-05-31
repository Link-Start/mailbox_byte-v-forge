package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mailboxapi/pb"
)

func upsertOutlookMailboxData(ctx context.Context, tx pgx.Tx, mailbox *pb.EmailMailbox, now int64) error {
	authStatus := strings.TrimSpace(mailbox.GetAuthStatus())
	if authStatus == "" {
		authStatus = authStatusOAuthPending
		if strings.TrimSpace(mailbox.GetRefreshToken()) != "" {
			authStatus = authStatusAuthorized
		}
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO mailbox_outlook_accounts (
			mailbox_email, password, refresh_token, access_token,
			auth_status, last_error, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$7)
		ON CONFLICT (mailbox_email) DO UPDATE SET
			password = CASE WHEN EXCLUDED.password <> '' THEN EXCLUDED.password ELSE mailbox_outlook_accounts.password END,
			refresh_token = CASE WHEN EXCLUDED.refresh_token <> '' THEN EXCLUDED.refresh_token ELSE mailbox_outlook_accounts.refresh_token END,
			access_token = CASE WHEN EXCLUDED.access_token <> '' THEN EXCLUDED.access_token ELSE mailbox_outlook_accounts.access_token END,
			auth_status = CASE
				WHEN $8 <> '' THEN EXCLUDED.auth_status
				WHEN EXCLUDED.refresh_token <> '' THEN 'AUTHORIZED'
				ELSE mailbox_outlook_accounts.auth_status
			END,
			last_error = CASE WHEN $8 <> '' OR EXCLUDED.last_error <> '' THEN EXCLUDED.last_error ELSE mailbox_outlook_accounts.last_error END,
			updated_at = EXCLUDED.updated_at
	`, emailx.Normalize(mailbox.GetEmailAddress()), strings.TrimSpace(mailbox.GetPassword()),
		strings.TrimSpace(mailbox.GetRefreshToken()), strings.TrimSpace(mailbox.GetAccessToken()),
		authStatus, strings.TrimSpace(mailbox.GetLastError()), now, strings.TrimSpace(mailbox.GetAuthStatus()))
	return err
}

func validateOutlookPollableMailbox(row *mailboxRow) error {
	if strings.TrimSpace(row.RefreshToken) == "" {
		return fmt.Errorf("mailbox has no refresh token: %s", emailx.Redact(row.Email))
	}
	if row.AuthStatus != authStatusAuthorized {
		return fmt.Errorf("mailbox is not authorized: %s auth_status=%s", emailx.Redact(row.Email), row.AuthStatus)
	}
	return nil
}

func updateOutlookAuthStatus(ctx context.Context, tx pgx.Tx, email string, authStatus string, lastError string, now int64) error {
	_, err := tx.Exec(ctx, `
		UPDATE mailbox_outlook_accounts
		SET auth_status = $1, last_error = $2, updated_at = $3
		WHERE mailbox_email = $4
	`, strings.TrimSpace(authStatus), strings.TrimSpace(lastError), now, emailx.Normalize(email))
	return err
}

func updateOutlookTokens(ctx context.Context, pool *pgxpool.Pool, email string, refreshToken string, accessToken string) error {
	_, err := pool.Exec(ctx, `
		UPDATE mailbox_outlook_accounts
		SET refresh_token = $1, access_token = $2, auth_status = $3, last_error = '', updated_at = $4
		WHERE mailbox_email = $5
	`, strings.TrimSpace(refreshToken), strings.TrimSpace(accessToken), authStatusAuthorized, time.Now().Unix(), emailx.Normalize(email))
	return err
}
