package main

import (
	"context"
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
)

type mailboxOperationRow struct {
	OperationID  string `gorm:"primaryKey;column:operation_id"`
	Action       string `gorm:"index"`
	Status       string `gorm:"index"`
	EmailAddress string `gorm:"index"`
	LastStep     string
	ErrorMessage string
	ImportOnly   bool
	OnlyMissing  bool
	Limit        int32
	ClaimOwner   string `gorm:"index"`
	ClaimUntil   int64  `gorm:"index"`
	AttemptCount int32
	ExitCode     int32
	MailboxCount int32
	FetchedCount int32
	FailedCount  int32
	MessageCount int32
	CreatedAt    int64 `gorm:"autoCreateTime"`
	UpdatedAt    int64 `gorm:"autoUpdateTime"`
}

func (mailboxOperationRow) TableName() string {
	return "mailbox_operations"
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
	db *gorm.DB
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

func newPgOperationStore(dsn string) (*pgOperationStore, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if !db.Migrator().HasTable((&mailboxOperationRow{}).TableName()) {
		return nil, errors.New("database schema is not migrated: missing table mailbox_operations")
	}
	return &pgOperationStore{db: db}, nil
}
