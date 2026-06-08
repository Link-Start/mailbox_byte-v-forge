package inboxapp

import (
	"context"
	"time"

	commonv1 "mailboxapi/internal/contracts/commonv1"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

type MessageRow struct {
	ID             string
	MailboxEmail   string
	Subject        string
	FromAddress    string
	BodyPreview    string
	ReceivedAtUnix int64
	RecipientsJSON string
	Provider       string
	SourceEmail    string
	BodyText       string
	HTMLBody       string
	RawSize        int64
}

type RecordMessagesRequest struct {
	Provider         string
	Messages         []*mailboxv1.EmailInboxMessage
	ExpandRecipients bool
	OutboxTable      string
	EventSource      string
	PrepareUnseen    func(context.Context, *mailboxv1.EmailInboxMessage) error
}

type Repository interface {
	InboxWatermark(ctx context.Context, email string) (int64, error)
	HasInboxMessages(ctx context.Context, email string) (bool, error)
	ListInboxRows(ctx context.Context, email string, limit int, receivedAfterUnix int64) ([]MessageRow, error)
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

type Service struct {
	repo        Repository
	providers   ProviderResolver
	recent      RecentCache
	secrets     SecretStore
	secretTTL   time.Duration
	outboxTable string
	eventSource string
	logf        func(string, ...any)
	now         func() time.Time
}

type Config struct {
	Repository  Repository
	Providers   ProviderResolver
	Recent      RecentCache
	Secrets     SecretStore
	SecretTTL   time.Duration
	OutboxTable string
	EventSource string
	Logf        func(string, ...any)
	Now         func() time.Time
}

func NewService(config Config) *Service {
	now := config.Now
	if now == nil {
		now = time.Now
	}
	return &Service{
		repo:        config.Repository,
		providers:   config.Providers,
		recent:      config.Recent,
		secrets:     config.Secrets,
		secretTTL:   config.SecretTTL,
		outboxTable: config.OutboxTable,
		eventSource: config.EventSource,
		logf:        config.Logf,
		now:         now,
	}
}

func (s *Service) log(format string, args ...any) {
	if s != nil && s.logf != nil {
		s.logf(format, args...)
	}
}

func (s *Service) normalizeProvider(provider string) string {
	if s != nil && s.providers != nil {
		return s.providers.NormalizeProviderInput(provider)
	}
	return NormalizeProviderKey(provider)
}
