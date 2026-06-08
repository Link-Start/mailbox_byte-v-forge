package main

import (
	"context"
	"errors"
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
	errOperationAlreadyExists  = errors.New("mailbox operation already exists")
	errOperationInvalidAction  = errors.New("mailbox operation action mismatch")
	errOperationIDRequired     = errors.New("operation_id is required")
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
