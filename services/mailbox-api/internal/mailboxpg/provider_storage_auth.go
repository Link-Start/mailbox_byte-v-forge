package mailboxpg

import (
	"fmt"
	"strings"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

func providerMailboxAuthStatus(mailbox *mailboxmodel.Record) (string, string) {
	authStatus := strings.TrimSpace(mailbox.GetAuthStatus())
	explicitAuthStatus := authStatus
	if authStatus != "" {
		return authStatus, explicitAuthStatus
	}
	if strings.TrimSpace(mailbox.GetRefreshToken()) != "" {
		return mailboxmodel.AuthStatusAuthorized, explicitAuthStatus
	}
	return mailboxmodel.AuthStatusOAuthPending, explicitAuthStatus
}

func providerStorageAuthUpdates(
	builder *providerStorageBuilder,
	fields mailboxprovider.TokenFields,
	authColumn string,
	lastErrorColumn string,
	explicitAuthStatus string,
) ([]any, error) {
	if authColumn == "" {
		return nil, nil
	}
	extraArgs := []any{explicitAuthStatus}
	explicitAuthArg := len(builder.args) + len(extraArgs)
	refreshColumn, err := mailboxprovider.SQLIdentifier(fields.RefreshTokenColumn)
	if err != nil {
		return nil, err
	}
	builder.setUpdate(authColumn, fmt.Sprintf(
		"%s = CASE WHEN $%d <> '' THEN EXCLUDED.%s WHEN EXCLUDED.%s <> '' THEN '%s' ELSE %s.%s END",
		authColumn, explicitAuthArg, authColumn, refreshColumn, mailboxmodel.AuthStatusAuthorized, builder.table, authColumn,
	))
	providerStorageLastErrorUpdate(builder, lastErrorColumn, explicitAuthArg)
	return extraArgs, nil
}

func providerStorageLastErrorUpdate(builder *providerStorageBuilder, lastErrorColumn string, explicitAuthArg int) {
	if lastErrorColumn == "" {
		return
	}
	builder.setUpdate(lastErrorColumn, fmt.Sprintf(
		"%s = CASE WHEN $%d <> '' OR EXCLUDED.%s <> '' THEN EXCLUDED.%s ELSE %s.%s END",
		lastErrorColumn, explicitAuthArg, lastErrorColumn, lastErrorColumn, builder.table, lastErrorColumn,
	))
}
