package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mailboxapi/pb"
)

func mailboxProviderUpsert(ctx context.Context, tx pgx.Tx, provider string, mailbox *pb.EmailMailbox, now int64) error {
	if definition := providerByKey(provider); definition != nil && definition.upsert != nil {
		return definition.upsert(ctx, tx, mailbox, now)
	}
	return nil
}

func mailboxProviderAuthFilter(provider string, authStatus string, args *[]any) string {
	definition := providerByKey(provider)
	if definition != nil {
		if definition.authFilter == nil {
			return "FALSE"
		}
		return definition.authFilter(authStatus, args)
	}
	parts := []string{}
	for _, definition := range mailboxProviderPlugins() {
		if definition.authFilter != nil {
			parts = append(parts, definition.authFilter(authStatus, args))
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
	definition := providerByKey(row.Provider)
	if definition == nil || definition.validatePoll == nil {
		return fmt.Errorf("mailbox provider cannot poll inbox: %s", row.Provider)
	}
	return definition.validatePoll(row)
}

func mailboxProviderUpdateAuth(ctx context.Context, tx pgx.Tx, provider string, email string, authStatus string, lastError string, now int64) error {
	if definition := providerByKey(provider); definition != nil && definition.updateAuth != nil {
		return definition.updateAuth(ctx, tx, email, authStatus, lastError, now)
	}
	return fmt.Errorf("mailbox provider has no auth state: %s", provider)
}

func mailboxProviderUpdateTokens(ctx context.Context, pool *pgxpool.Pool, provider string, email string, refreshToken string, accessToken string) error {
	if definition := providerByKey(provider); definition != nil && definition.updateTokens != nil {
		return definition.updateTokens(ctx, pool, email, refreshToken, accessToken)
	}
	return fmt.Errorf("mailbox provider has no token storage: %s", provider)
}

func mailboxProviderPruneInbound(ctx context.Context, tx pgx.Tx, provider string, retention mailboxInboxRetention) error {
	if definition := providerByKey(provider); definition != nil && definition.pruneInbound != nil {
		return definition.pruneInbound(ctx, tx, retention)
	}
	return nil
}

func listMailboxProviderVirtualMailboxes(ctx context.Context, pool *pgxpool.Pool, query mailboxListQuery) ([]*pb.EmailMailbox, error) {
	out := []*pb.EmailMailbox{}
	for _, definition := range mailboxProviderPlugins() {
		if query.Provider != "" && query.Provider != definition.key {
			continue
		}
		if definition.virtualMailboxes == nil {
			continue
		}
		if definition.includeVirtual != nil && !definition.includeVirtual(query.AuthStatus) {
			continue
		}
		items, err := definition.virtualMailboxes(ctx, pool, query)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

func prepareMailboxProjection(mailbox *pb.EmailMailbox) {
	if mailbox == nil {
		return
	}
	if definition := providerByKey(mailbox.GetProviderKey()); definition != nil && definition.prepareProjection != nil {
		definition.prepareProjection(mailbox)
	}
}
