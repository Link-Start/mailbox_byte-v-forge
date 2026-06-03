package main

import (
	"context"
	"errors"
	"strings"
	"time"

	commonv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/common/v1"
	"github.com/byte-v-forge/common-lib/redisx"
	"github.com/byte-v-forge/common-lib/secretref"
	"github.com/redis/go-redis/v9"
)

type mailboxSecretStore struct {
	store *redisx.StringStore
	ttl   time.Duration
}

func newMailboxSecretStore(client redis.Cmdable, prefix string, ttl time.Duration) *mailboxSecretStore {
	return &mailboxSecretStore{store: redisx.NewStringStore(client, prefix, ttl), ttl: ttl}
}

func (s *mailboxSecretStore) SaveOTP(ctx context.Context, ref *commonv1.SecretRef, value string, expiresAt time.Time) error {
	if s == nil || s.store == nil {
		return errors.New("mailbox secret store is not configured")
	}
	if err := secretref.Validate(ref); err != nil {
		return err
	}
	if strings.TrimSpace(value) == "" {
		return errors.New("otp value is required")
	}
	ttl := s.ttl
	if !expiresAt.IsZero() {
		ttl = time.Until(expiresAt)
	}
	if ttl <= 0 {
		return nil
	}
	return s.store.SaveTTL(ctx, ref.GetSecretId(), value, ttl)
}

func (s *mailboxSecretStore) ResolveSecret(ctx context.Context, ref *commonv1.SecretRef) (string, error) {
	if s == nil || s.store == nil {
		return "", errors.New("mailbox secret store is not configured")
	}
	if err := secretref.Validate(ref); err != nil {
		return "", err
	}
	if strings.TrimSpace(ref.GetProvider()) != "mailbox" || strings.TrimSpace(ref.GetPurpose()) != "email_otp" {
		return "", errors.New("mailbox secret ref scope mismatch")
	}
	value, ok, err := s.store.Load(ctx, ref.GetSecretId())
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("mailbox secret ref is not resolvable")
	}
	return value, nil
}
