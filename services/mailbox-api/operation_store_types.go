package main

import mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

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
