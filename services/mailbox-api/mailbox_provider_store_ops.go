package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

func (s *MailboxStore) mailboxProviderUpsert(ctx context.Context, tx pgx.Tx, provider string, mailbox *mailboxmodel.Record, now int64) error {
	if definition := s.providerStorageByKey(provider); definition != nil {
		return definition.Upsert(ctx, tx, mailbox, now)
	}
	return nil
}

func (s *MailboxStore) mailboxProviderAuthFilter(provider string, authStatus string, args *[]any) string {
	definition := s.providerStorageByKey(provider)
	if definition != nil {
		filter, ok := definition.AuthFilter(authStatus, args)
		if !ok {
			return "FALSE"
		}
		return filter
	}
	parts := []string{}
	for _, definition := range s.mailboxProviderStorageExtensions() {
		if filter, ok := definition.AuthFilter(authStatus, args); ok {
			parts = append(parts, filter)
		}
	}
	if len(parts) == 0 {
		return "FALSE"
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func (s *MailboxStore) mailboxProviderValidatePoll(row *mailboxRow) error {
	if row == nil {
		return fmt.Errorf("mailbox is required")
	}
	definition := s.providerStorageByKey(row.Provider)
	if definition == nil || !definition.CanValidatePoll() {
		return fmt.Errorf("mailbox provider cannot poll inbox: %s", row.Provider)
	}
	return definition.ValidatePoll(mailboxprovider.MailboxRecord{
		Email:        row.Email,
		Provider:     row.Provider,
		RefreshToken: row.RefreshToken,
		AuthStatus:   row.AuthStatus,
	})
}

func (s *MailboxStore) mailboxProviderUpdateAuth(ctx context.Context, tx pgx.Tx, provider string, email string, authStatus string, lastError string, now int64) error {
	if definition := s.providerStorageByKey(provider); definition != nil && definition.CanUpdateAuth() {
		return definition.UpdateAuth(ctx, tx, email, authStatus, lastError, now)
	}
	return fmt.Errorf("mailbox provider has no auth state: %s", provider)
}

func (s *MailboxStore) mailboxProviderUpdateTokens(ctx context.Context, pool *pgxpool.Pool, provider string, email string, refreshToken string, accessToken string) error {
	if definition := s.providerStorageByKey(provider); definition != nil && definition.CanUpdateTokens() {
		return definition.UpdateTokens(ctx, pool, email, refreshToken, accessToken)
	}
	return fmt.Errorf("mailbox provider has no token storage: %s", provider)
}

func (s *MailboxStore) mailboxProviderPruneInbound(ctx context.Context, tx pgx.Tx, provider string, retention mailboxprovider.InboxRetention) error {
	if definition := s.providerRetentionByKey(provider); definition != nil {
		return definition.PruneInbound(ctx, tx, retention)
	}
	return nil
}

func (s *MailboxStore) listMailboxProviderVirtualMailboxes(ctx context.Context, pool *pgxpool.Pool, query mailboxprovider.ListQuery) ([]*mailboxmodel.Record, error) {
	out := []*mailboxmodel.Record{}
	for _, definition := range s.mailboxProviderVirtualSources() {
		if query.Provider != "" && query.Provider != definition.Key() {
			continue
		}
		if !definition.HasVirtualMailboxes() {
			continue
		}
		if !definition.IncludeVirtual(query.AuthStatus) {
			continue
		}
		items, err := definition.VirtualMailboxes(ctx, pool, query)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

func (s *MailboxStore) providerStorageByKey(provider string) mailboxprovider.StorageExtension {
	if s == nil || s.providers == nil {
		return nil
	}
	if plugin := s.providers.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func (s *MailboxStore) providerRetentionByKey(provider string) mailboxprovider.InboxRetentionPolicy {
	if s == nil || s.providers == nil {
		return nil
	}
	if plugin := s.providers.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func (s *MailboxStore) capabilityProviderByKey(provider string) mailboxprovider.CapabilityPlugin {
	if s == nil || s.providers == nil {
		return nil
	}
	if plugin := s.providers.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func (s *MailboxStore) mailboxProviderStorageExtensions() []mailboxprovider.StorageExtension {
	if s == nil || s.providers == nil {
		return nil
	}
	return s.providers.StorageExtensions()
}

func (s *MailboxStore) mailboxProviderVirtualSources() []mailboxprovider.VirtualMailboxSource {
	if s == nil || s.providers == nil {
		return nil
	}
	return s.providers.VirtualMailboxSources()
}

func (s *MailboxStore) defaultMailboxProvider() string {
	if s == nil || s.providers == nil {
		return ""
	}
	return s.providers.DefaultKey()
}

func (s *MailboxStore) normalizeMailboxProviderInput(provider string) string {
	value := mailboxprovider.NormalizeKey(provider)
	if value == "" {
		return ""
	}
	if s != nil && s.providers != nil {
		if definition := s.providers.ByKey(value); definition != nil {
			return definition.Key()
		}
	}
	return value
}

func (s *MailboxStore) prepareMailboxProjection(mailbox *mailboxmodel.Record) {
	if mailbox == nil {
		return
	}
	if definition := s.capabilityProviderByKey(mailbox.GetProviderKey()); definition != nil {
		definition.PrepareProjection(mailbox)
	}
}

func prepareMailboxProjection(mailbox *mailboxmodel.Record) {
	if mailbox == nil {
		return
	}
	if definition := capabilityProviderByKey(mailbox.GetProviderKey()); definition != nil {
		definition.PrepareProjection(mailbox)
	}
}
