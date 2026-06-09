package mailboxpg

import (
	"context"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

func (r *Repository) listVirtualMailboxes(ctx context.Context, query mailboxprovider.ListQuery) ([]*mailboxmodel.Record, error) {
	out := []*mailboxmodel.Record{}
	for _, source := range r.providers.VirtualMailboxSources() {
		if query.Provider != "" && query.Provider != source.Key() {
			continue
		}
		if !source.StoredInboxOnly() || !source.IncludeVirtual(query.AuthStatus) {
			continue
		}
		items, err := r.listStoredInboxOnlyVirtualMailboxes(ctx, source.Key(), query, r.providers.NormalizeProviderInput, r.providers.PrepareProjection)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}
