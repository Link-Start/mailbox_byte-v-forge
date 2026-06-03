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

func mailboxProviderUpsert(ctx context.Context, tx pgx.Tx, provider string, mailbox *mailboxmodel.Record, now int64) error {
	if definition := providerStorageByKey(provider); definition != nil {
		return definition.Upsert(ctx, tx, mailbox, now)
	}
	return nil
}

func mailboxProviderAuthFilter(provider string, authStatus string, args *[]any) string {
	definition := providerStorageByKey(provider)
	if definition != nil {
		filter, ok := definition.AuthFilter(authStatus, args)
		if !ok {
			return "FALSE"
		}
		return filter
	}
	parts := []string{}
	for _, definition := range mailboxProviderStorageExtensions() {
		if filter, ok := definition.AuthFilter(authStatus, args); ok {
			parts = append(parts, filter)
		}
	}
	if len(parts) == 0 {
		return "FALSE"
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func mailboxProviderValidatePoll(row *mailboxRow) error {
	if row == nil {
		return fmt.Errorf("mailbox is required")
	}
	definition := providerStorageByKey(row.Provider)
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

func mailboxProviderUpdateAuth(ctx context.Context, tx pgx.Tx, provider string, email string, authStatus string, lastError string, now int64) error {
	if definition := providerStorageByKey(provider); definition != nil && definition.CanUpdateAuth() {
		return definition.UpdateAuth(ctx, tx, email, authStatus, lastError, now)
	}
	return fmt.Errorf("mailbox provider has no auth state: %s", provider)
}

func mailboxProviderUpdateTokens(ctx context.Context, pool *pgxpool.Pool, provider string, email string, refreshToken string, accessToken string) error {
	if definition := providerStorageByKey(provider); definition != nil && definition.CanUpdateTokens() {
		return definition.UpdateTokens(ctx, pool, email, refreshToken, accessToken)
	}
	return fmt.Errorf("mailbox provider has no token storage: %s", provider)
}

func mailboxProviderPruneInbound(ctx context.Context, tx pgx.Tx, provider string, retention mailboxprovider.InboxRetention) error {
	if definition := providerRetentionByKey(provider); definition != nil {
		return definition.PruneInbound(ctx, tx, retention)
	}
	return nil
}

func listMailboxProviderVirtualMailboxes(ctx context.Context, pool *pgxpool.Pool, query mailboxprovider.ListQuery) ([]*mailboxmodel.Record, error) {
	out := []*mailboxmodel.Record{}
	for _, definition := range mailboxProviderVirtualSources() {
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

func prepareMailboxProjection(mailbox *mailboxmodel.Record) {
	if mailbox == nil {
		return
	}
	if definition := capabilityProviderByKey(mailbox.GetProviderKey()); definition != nil {
		definition.PrepareProjection(mailbox)
	}
}
