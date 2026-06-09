package mailboxpg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/mailboxprovider"
)

type inboxMessageKey struct {
	provider     string
	mailboxEmail string
	messageKey   string
}

func (r *Repository) pruneInbound(ctx context.Context, tx pgx.Tx, provider string, retention mailboxprovider.InboxRetention) error {
	definition := r.providers.RetentionByKey(provider)
	if definition == nil {
		return nil
	}
	policy, ok := definition.RetentionPolicy()
	if !ok {
		return nil
	}
	switch policy.Scope {
	case mailboxprovider.RetentionScopeDomain:
		for domain := range retention.TouchedDomains {
			if err := pruneDomainMessages(ctx, tx, provider, domain, policy.MaxMessages); err != nil {
				return err
			}
		}
	case mailboxprovider.RetentionScopeMailbox:
		for mailboxEmail := range retention.TouchedMailboxes {
			if err := pruneMailboxMessages(ctx, tx, provider, mailboxEmail, policy.MaxMessages); err != nil {
				return err
			}
		}
	}
	return nil
}
