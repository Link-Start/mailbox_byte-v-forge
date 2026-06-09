package inboxapp

import "time"

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
