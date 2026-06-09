package inboxapp

import (
	"context"
	"time"

	commonv1 "mailboxapi/internal/contracts/commonv1"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

type Repository interface {
	InboxWatermark(ctx context.Context, email string) (int64, error)
	HasInboxMessages(ctx context.Context, email string) (bool, error)
	ListInboxRows(ctx context.Context, email string, limit int, receivedAfterUnix int64) ([]MessageRow, error)
	GetInboxRow(ctx context.Context, email string, messageID string, provider string) (MessageRow, bool, error)
	LatestInboxRows(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, limit int) ([]MessageRow, error)
	RecordMessages(ctx context.Context, request RecordMessagesRequest) ([]*mailboxv1.EmailInboxMessage, error)
}

type ProviderResolver interface {
	NormalizeProviderInput(provider string) string
}

type RecentCache interface {
	Record(ctx context.Context, messages []*mailboxv1.EmailInboxMessage) error
	Latest(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, parserProfile string, signalKind mailboxv1.EmailSignalKind) (*mailboxv1.EmailInboxMessage, bool, error)
}

type SecretStore interface {
	SaveOTP(ctx context.Context, ref *commonv1.SecretRef, code string, expiresAt time.Time) error
}
