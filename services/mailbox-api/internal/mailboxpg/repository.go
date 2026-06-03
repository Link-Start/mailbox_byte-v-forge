package mailboxpg

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/accountmodel"
	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/byte-v-forge/common-lib/pagex"
	"github.com/byte-v-forge/common-lib/randx"
	"github.com/byte-v-forge/common-lib/redactx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

const mailboxErrorSnippetLimit = 600

type Repository struct {
	pool      *pgxpool.Pool
	providers *mailboxprovider.Registry
}

func NewRepository(pool *pgxpool.Pool, providers *mailboxprovider.Registry) (*Repository, error) {
	if pool == nil {
		return nil, errors.New("mailbox pg pool is required")
	}
	if providers == nil {
		return nil, errors.New("mailbox providers are required")
	}
	return &Repository{pool: pool, providers: providers}, nil
}

func (r *Repository) UpsertMailbox(ctx context.Context, mailbox *mailboxmodel.Record) (*mailboxmodel.Record, error) {
	if mailbox == nil {
		return nil, errors.New("mailbox is required")
	}
	email := emailx.Normalize(mailbox.GetEmailAddress())
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	requestedProvider := r.providers.NormalizeProviderInput(mailbox.GetProviderKey())
	insertProvider := requestedProvider
	if insertProvider == "" {
		insertProvider = r.providers.DefaultKey()
	}
	rowID, err := randx.Hex(16)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var persistedProvider string
	if err := tx.QueryRow(ctx, `
		INSERT INTO mailboxes (id, email, provider, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$4)
		ON CONFLICT (email) DO UPDATE SET
			provider = CASE WHEN $5 <> '' THEN EXCLUDED.provider ELSE mailboxes.provider END,
			updated_at = EXCLUDED.updated_at
		RETURNING provider
	`, rowID, email, insertProvider, now, requestedProvider).Scan(&persistedProvider); err != nil {
		return nil, err
	}
	if err := r.providers.Upsert(ctx, tx, persistedProvider, mailbox, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindMailbox(ctx, email)
}

func (r *Repository) MarkEmailAuthStatus(ctx context.Context, email string, authStatus string, lastError string) (*mailboxmodel.Record, error) {
	email = emailx.Normalize(email)
	authStatus = strings.TrimSpace(authStatus)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	if authStatus == "" {
		return nil, errors.New("auth_status is required")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row, err := ScanMailbox(tx.QueryRow(ctx, r.providers.MailboxSelectSQL()+" WHERE m.email = $1 FOR UPDATE", email))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	if err := r.providers.UpdateAuth(ctx, tx, row.Provider, email, authStatus, safeText(lastError), now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, "UPDATE mailboxes SET updated_at = $1 WHERE email = $2", now, email); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindMailbox(ctx, email)
}

func (r *Repository) FindMailbox(ctx context.Context, email string) (*mailboxmodel.Record, error) {
	row, err := ScanMailbox(r.pool.QueryRow(ctx, r.providers.MailboxSelectSQL()+" WHERE m.email = $1", emailx.Normalize(email)))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	if err != nil {
		return nil, err
	}
	return r.recordFromRow(row), nil
}

func (r *Repository) PollMailboxForEmail(ctx context.Context, email string) (*mailboxmodel.Record, error) {
	email = emailx.Normalize(email)
	row, err := ScanMailbox(r.pool.QueryRow(ctx, r.providers.MailboxSelectSQL()+" WHERE m.email = $1", email))
	if errors.Is(err, pgx.ErrNoRows) {
		canonical := emailx.CanonicalPlusAlias(email)
		if canonical != "" && canonical != email {
			row, err = ScanMailbox(r.pool.QueryRow(ctx, r.providers.MailboxSelectSQL()+" WHERE m.email = $1", canonical))
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	if err != nil {
		return nil, err
	}
	if err := r.providers.ValidatePoll(row.ToProviderRecord()); err != nil {
		return nil, err
	}
	return r.recordFromRow(row), nil
}

func (r *Repository) UpdateMailboxTokens(ctx context.Context, email string, refreshToken string, accessToken string) error {
	email = emailx.Normalize(email)
	row, err := ScanMailbox(r.pool.QueryRow(ctx, r.providers.MailboxSelectSQL()+" WHERE m.email = $1", email))
	if err != nil {
		return err
	}
	if err := r.providers.UpdateTokens(ctx, r.pool, row.Provider, email, refreshToken, accessToken); err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, "UPDATE mailboxes SET updated_at = $1 WHERE email = $2", time.Now().Unix(), email)
	return err
}

func (r *Repository) ListMailboxes(ctx context.Context, authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxmodel.ListPage, error) {
	query, err := r.newMailboxListQuery(authStatus, provider, emailAddress, cursorValue, limit)
	if err != nil {
		return mailboxmodel.ListPage{}, err
	}
	stored, err := r.listStoredMailboxes(ctx, query)
	if err != nil {
		return mailboxmodel.ListPage{}, err
	}
	rows := stored
	virtual, err := r.providers.VirtualMailboxes(ctx, r.pool, query)
	if err != nil {
		return mailboxmodel.ListPage{}, err
	}
	if len(virtual) > 0 {
		rows = append(rows, virtual...)
	}
	return mailboxPageFromRows(rows, query.Limit), nil
}

func (r *Repository) ListOAuthMailboxes(ctx context.Context, limit int32) ([]*mailboxmodel.Record, error) {
	n := int(limit)
	if n <= 0 {
		n = 100
	}
	if n > 500 {
		n = 500
	}
	args := []any{}
	query := r.providers.MailboxSelectSQL() + " WHERE " + r.providers.AuthFilter("", mailboxmodel.AuthStatusAuthorized, &args)
	args = append(args, n)
	query += fmt.Sprintf(" ORDER BY m.updated_at DESC, m.email DESC LIMIT $%d", len(args))
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*mailboxmodel.Record{}
	for rows.Next() {
		row, err := ScanMailbox(rows)
		if err != nil {
			return nil, err
		}
		if r.providers.ValidatePoll(row.ToProviderRecord()) != nil {
			continue
		}
		out = append(out, r.recordFromRow(row))
	}
	return out, rows.Err()
}

func (r *Repository) DeleteMailbox(ctx context.Context, email string) (bool, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return false, errors.New("email_address is required")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row, err := ScanMailbox(tx.QueryRow(ctx, r.providers.MailboxSelectSQL()+" WHERE m.email = $1 FOR UPDATE", email))
	if errors.Is(err, pgx.ErrNoRows) {
		deleted, deleteErr := deleteMailboxInbox(ctx, tx, []string{email})
		if deleteErr != nil {
			return false, deleteErr
		}
		if err := tx.Commit(ctx); err != nil {
			return false, err
		}
		return deleted, nil
	}
	if err != nil {
		return false, err
	}

	deleteEmails := []string{row.Email}
	if _, err := deleteMailboxInbox(ctx, tx, deleteEmails); err != nil {
		return false, err
	}
	args, inClause := sqlInArgs(deleteEmails)
	tag, err := tx.Exec(ctx, "DELETE FROM mailboxes WHERE email IN ("+inClause+")", args...)
	if err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *Repository) newMailboxListQuery(authStatus string, provider string, emailAddress string, cursorValue string, limit int32) (mailboxprovider.ListQuery, error) {
	cursor, err := pagex.DecodeKeysetCursor(cursorValue)
	if err != nil {
		return mailboxprovider.ListQuery{}, mailboxmodel.ErrInvalidMailboxListCursor
	}
	return mailboxprovider.ListQuery{
		AuthStatus:   strings.TrimSpace(authStatus),
		Provider:     r.providers.NormalizeProviderInput(provider),
		EmailAddress: emailx.Normalize(emailAddress),
		Cursor:       cursor,
		Limit:        accountmodel.NormalizePageLimit(int(limit)),
	}, nil
}

func (r *Repository) listStoredMailboxes(ctx context.Context, filter mailboxprovider.ListQuery) ([]*mailboxmodel.Record, error) {
	args := []any{}
	query := r.providers.MailboxSelectSQL() + ` WHERE 1=1`
	if filter.AuthStatus != "" {
		query += " AND " + r.providers.AuthFilter(filter.Provider, filter.AuthStatus, &args)
	}
	if filter.Provider != "" {
		args = append(args, filter.Provider)
		query += fmt.Sprintf(" AND m.provider = $%d", len(args))
	}
	if filter.EmailAddress != "" {
		args = append(args, filter.EmailAddress)
		query += fmt.Sprintf(" AND m.email = $%d", len(args))
	}
	if filter.HasCursor() {
		args = append(args, filter.Cursor.UpdatedAt.Unix(), emailx.Normalize(filter.Cursor.ID))
		query += fmt.Sprintf(" AND (m.updated_at < $%d OR (m.updated_at = $%d AND m.email < $%d))", len(args)-1, len(args)-1, len(args))
	}
	args = append(args, filter.ScanLimit())
	query += fmt.Sprintf(" ORDER BY m.updated_at DESC, m.email DESC LIMIT $%d", len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*mailboxmodel.Record{}
	for rows.Next() {
		row, err := ScanMailbox(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r.recordFromRow(row))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) recordFromRow(row *MailboxRow) *mailboxmodel.Record {
	return row.ToRecord(r.providers.NormalizeProviderInput, r.providers.PrepareProjection)
}

func mailboxPageFromRows(rows []*mailboxmodel.Record, limit int) mailboxmodel.ListPage {
	rows = uniqueMailboxRows(rows)
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].GetUpdatedAt() == rows[j].GetUpdatedAt() {
			return rows[i].GetEmailAddress() > rows[j].GetEmailAddress()
		}
		return rows[i].GetUpdatedAt() > rows[j].GetUpdatedAt()
	})
	page := pagex.NewKeysetPage(rows, limit, func(mailbox *mailboxmodel.Record) pagex.KeysetCursor {
		return pagex.KeysetCursorValue(time.Unix(mailbox.GetUpdatedAt(), 0).UTC(), mailbox.GetEmailAddress())
	})
	return mailboxmodel.ListPage{Mailboxes: page.Items, NextCursor: page.NextCursor}
}

func uniqueMailboxRows(rows []*mailboxmodel.Record) []*mailboxmodel.Record {
	seen := map[string]struct{}{}
	out := make([]*mailboxmodel.Record, 0, len(rows))
	for _, row := range rows {
		email := emailx.Normalize(row.GetEmailAddress())
		if email == "" {
			continue
		}
		if _, exists := seen[email]; exists {
			continue
		}
		seen[email] = struct{}{}
		out = append(out, row)
	}
	return out
}

func deleteMailboxInbox(ctx context.Context, tx pgx.Tx, emails []string) (bool, error) {
	args, inClause := sqlInArgs(emails)
	messageTag, err := tx.Exec(ctx, "DELETE FROM mailbox_inbox_messages WHERE mailbox_email IN ("+inClause+")", args...)
	if err != nil {
		return false, err
	}
	seenTag, err := tx.Exec(ctx, "DELETE FROM mailbox_inbox_seen WHERE mailbox_email IN ("+inClause+")", args...)
	if err != nil {
		return false, err
	}
	return messageTag.RowsAffected() > 0 || seenTag.RowsAffected() > 0, nil
}

func sqlInArgs(values []string) ([]any, string) {
	args := make([]any, 0, len(values))
	placeholders := make([]string, 0, len(values))
	for _, item := range values {
		args = append(args, item)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
	}
	return args, strings.Join(placeholders, ",")
}

func safeText(value string) string {
	return redactx.TextSnippet(value, mailboxErrorSnippetLimit)
}
