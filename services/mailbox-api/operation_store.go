package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

const (
	operationActionRegisterMailbox = "REGISTER_MAILBOX"
	operationActionMailboxOAuth    = "MAILBOX_OAUTH"
	operationActionFetchInboxes    = "FETCH_INBOXES"

	operationStatusCreated   = "CREATED"
	operationStatusRunning   = "RUNNING"
	operationStatusSucceeded = "SUCCEEDED"
	operationStatusFailed    = "FAILED"
)

const operationActionRunLeaseSeconds int32 = 2 * 60 * 60

var (
	errOperationAlreadyRunning = errors.New("mailbox operation is already running")
	errOperationInvalidAction  = errors.New("mailbox operation action mismatch")
	errOperationNotFound       = errors.New("mailbox operation not found")
)

type mailboxOperationRow struct {
	OperationID  string
	Action       string
	Status       string
	EmailAddress string
	LastStep     string
	ErrorMessage string
	ImportOnly   bool
	OnlyMissing  bool
	Limit        int32
	ClaimOwner   string
	ClaimUntil   int64
	AttemptCount int32
	ExitCode     int32
	MailboxCount int32
	FetchedCount int32
	FailedCount  int32
	MessageCount int32
	CreatedAt    int64
	UpdatedAt    int64
}

type operationStore interface {
	create(ctx context.Context, operationID, action, emailAddress string) (*mailboxv1.MailboxOperation, error)
	createRegistration(ctx context.Context, operationID string, importOnly bool) (*mailboxv1.MailboxOperation, error)
	createOAuth(ctx context.Context, operationID string, emailAddress string, onlyMissing bool, limit int32) (*mailboxv1.MailboxOperation, error)
	update(ctx context.Context, operationID string, update operationUpdate) (*mailboxv1.MailboxOperation, error)
	get(ctx context.Context, operationID string) (*mailboxv1.MailboxOperation, error)
	list(ctx context.Context, filter operationListFilter) ([]*mailboxv1.MailboxOperation, error)
	startRegistrationWorkerRun(ctx context.Context, operationID string) (*operationRunStart, error)
	startOAuthWorkerRun(ctx context.Context, operationID string) (*operationRunStart, error)
}

type pgOperationStore struct {
	pool *pgxpool.Pool
}

type operationUpdate struct {
	Status       string
	LastStep     string
	ErrorMessage string
	ExitCode     int32
	MailboxCount int32
	FetchedCount int32
	FailedCount  int32
	MessageCount int32
}

type operationListFilter struct {
	Limit        int
	Status       mailboxv1.MailboxOperationStatus
	Action       mailboxv1.MailboxOperationAction
	EmailAddress string
}

type operationRunStart struct {
	Operation    *mailboxv1.MailboxOperation
	EmailAddress string
	ImportOnly   bool
	OnlyMissing  bool
	Limit        int32
	Final        bool
}

func newPgOperationStore(ctx context.Context, dsn string) (*pgOperationStore, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("mailbox postgres DSN is required")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	store := &pgOperationStore{pool: pool}
	if err := store.ensureSchema(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (s *pgOperationStore) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *pgOperationStore) ensureSchema(ctx context.Context) error {
	for _, statement := range operationSchemaStatements() {
		if _, err := s.pool.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func operationSchemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS mailbox_operations (
			operation_id text PRIMARY KEY,
			action text NOT NULL DEFAULT '',
			status text NOT NULL DEFAULT '',
			email_address text NOT NULL DEFAULT '',
			last_step text NOT NULL DEFAULT '',
			error_message text NOT NULL DEFAULT '',
			import_only boolean NOT NULL DEFAULT false,
			only_missing boolean NOT NULL DEFAULT false,
			"limit" integer NOT NULL DEFAULT 0,
			claim_owner text NOT NULL DEFAULT '',
			claim_until bigint NOT NULL DEFAULT 0,
			attempt_count integer NOT NULL DEFAULT 0,
			exit_code integer NOT NULL DEFAULT 0,
			mailbox_count integer NOT NULL DEFAULT 0,
			fetched_count integer NOT NULL DEFAULT 0,
			failed_count integer NOT NULL DEFAULT 0,
			message_count integer NOT NULL DEFAULT 0,
			created_at bigint NOT NULL DEFAULT 0,
			updated_at bigint NOT NULL DEFAULT 0
		)`,
		`ALTER TABLE mailbox_operations ADD COLUMN IF NOT EXISTS import_only boolean NOT NULL DEFAULT false`,
		`ALTER TABLE mailbox_operations ADD COLUMN IF NOT EXISTS only_missing boolean NOT NULL DEFAULT false`,
		`ALTER TABLE mailbox_operations ADD COLUMN IF NOT EXISTS "limit" integer NOT NULL DEFAULT 0`,
		`ALTER TABLE mailbox_operations ADD COLUMN IF NOT EXISTS claim_owner text NOT NULL DEFAULT ''`,
		`ALTER TABLE mailbox_operations ADD COLUMN IF NOT EXISTS claim_until bigint NOT NULL DEFAULT 0`,
		`ALTER TABLE mailbox_operations ADD COLUMN IF NOT EXISTS attempt_count integer NOT NULL DEFAULT 0`,
		`ALTER TABLE mailbox_operations ADD COLUMN IF NOT EXISTS exit_code integer NOT NULL DEFAULT 0`,
		`ALTER TABLE mailbox_operations ADD COLUMN IF NOT EXISTS mailbox_count integer NOT NULL DEFAULT 0`,
		`ALTER TABLE mailbox_operations ADD COLUMN IF NOT EXISTS fetched_count integer NOT NULL DEFAULT 0`,
		`ALTER TABLE mailbox_operations ADD COLUMN IF NOT EXISTS failed_count integer NOT NULL DEFAULT 0`,
		`ALTER TABLE mailbox_operations ADD COLUMN IF NOT EXISTS message_count integer NOT NULL DEFAULT 0`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_operations_action ON mailbox_operations(action)`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_operations_status ON mailbox_operations(status)`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_operations_email_address ON mailbox_operations(email_address)`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_operations_claim_owner ON mailbox_operations(claim_owner)`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_operations_claim_until ON mailbox_operations(claim_until)`,
	}
}

type operationRowScanner interface {
	Scan(dest ...any) error
}

func scanOperationRow(scanner operationRowScanner) (mailboxOperationRow, error) {
	var row mailboxOperationRow
	err := scanner.Scan(
		&row.OperationID,
		&row.Action,
		&row.Status,
		&row.EmailAddress,
		&row.LastStep,
		&row.ErrorMessage,
		&row.ImportOnly,
		&row.OnlyMissing,
		&row.Limit,
		&row.ClaimOwner,
		&row.ClaimUntil,
		&row.AttemptCount,
		&row.ExitCode,
		&row.MailboxCount,
		&row.FetchedCount,
		&row.FailedCount,
		&row.MessageCount,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	return row, err
}

func operationSelectSQL() string {
	return fmt.Sprintf(`SELECT %s FROM mailbox_operations`, operationColumns())
}

func operationColumns() string {
	return `operation_id, action, status, email_address, last_step, error_message,
		import_only, only_missing, "limit", claim_owner, claim_until, attempt_count,
		exit_code, mailbox_count, fetched_count, failed_count, message_count, created_at, updated_at`
}
