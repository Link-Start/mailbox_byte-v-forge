package mailboxprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mailboxapi/internal/mailboxmodel"
)

func (r *Registry) NormalizeProviderInput(provider string) string {
	value := NormalizeKey(provider)
	if value == "" {
		return ""
	}
	if definition := r.ByKey(value); definition != nil {
		return definition.Key()
	}
	return value
}

func (r *Registry) Upsert(ctx context.Context, tx pgx.Tx, provider string, mailbox *mailboxmodel.Record, now int64) error {
	if definition := r.StorageByKey(provider); definition != nil {
		return definition.Upsert(ctx, tx, mailbox, now)
	}
	return nil
}

func (r *Registry) AuthFilter(provider string, authStatus string, args *[]any) string {
	definition := r.StorageByKey(provider)
	if definition != nil {
		filter, ok := definition.AuthFilter(authStatus, args)
		if !ok {
			return "FALSE"
		}
		return filter
	}
	parts := []string{}
	for _, definition := range r.StorageExtensions() {
		if filter, ok := definition.AuthFilter(authStatus, args); ok {
			parts = append(parts, filter)
		}
	}
	if len(parts) == 0 {
		return "FALSE"
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func (r *Registry) ValidatePoll(row MailboxRecord) error {
	if strings.TrimSpace(row.Email) == "" {
		return fmt.Errorf("mailbox is required")
	}
	definition := r.StorageByKey(row.Provider)
	if definition == nil || !definition.CanValidatePoll() {
		return fmt.Errorf("mailbox provider cannot poll inbox: %s", row.Provider)
	}
	return definition.ValidatePoll(row)
}

func (r *Registry) UpdateAuth(ctx context.Context, tx pgx.Tx, provider string, email string, authStatus string, lastError string, now int64) error {
	if definition := r.StorageByKey(provider); definition != nil && definition.CanUpdateAuth() {
		return definition.UpdateAuth(ctx, tx, email, authStatus, lastError, now)
	}
	return fmt.Errorf("mailbox provider has no auth state: %s", provider)
}

func (r *Registry) UpdateTokens(ctx context.Context, pool *pgxpool.Pool, provider string, email string, refreshToken string, accessToken string) error {
	if definition := r.StorageByKey(provider); definition != nil && definition.CanUpdateTokens() {
		return definition.UpdateTokens(ctx, pool, email, refreshToken, accessToken)
	}
	return fmt.Errorf("mailbox provider has no token storage: %s", provider)
}

func (r *Registry) PruneInbound(ctx context.Context, tx pgx.Tx, provider string, retention InboxRetention) error {
	if definition := r.RetentionByKey(provider); definition != nil {
		return definition.PruneInbound(ctx, tx, retention)
	}
	return nil
}

func (r *Registry) VirtualMailboxes(ctx context.Context, pool *pgxpool.Pool, query ListQuery) ([]*mailboxmodel.Record, error) {
	out := []*mailboxmodel.Record{}
	for _, definition := range r.VirtualMailboxSources() {
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

func (r *Registry) PrepareProjection(mailbox *mailboxmodel.Record) {
	if mailbox == nil {
		return
	}
	if definition := r.CapabilityByKey(mailbox.GetProviderKey()); definition != nil {
		definition.PrepareProjection(mailbox)
	}
}

func (r *Registry) StorageByKey(provider string) StorageExtension {
	if plugin := r.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func (r *Registry) RetentionByKey(provider string) InboxRetentionPolicy {
	if plugin := r.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func (r *Registry) CapabilityByKey(provider string) CapabilityPlugin {
	if plugin := r.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}
