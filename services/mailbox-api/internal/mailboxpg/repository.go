package mailboxpg

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"mailboxapi/internal/redactx"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

const mailboxErrorSnippetLimit = 600

type Repository struct {
	pool      *pgxpool.Pool
	providers *mailboxprovider.Registry
}

func NewRepository(pool *pgxpool.Pool, providers *mailboxprovider.Registry) (*Repository, error) {
	if pool == nil {
		return nil, errors.New("mailbox pg pool is required")
	}
	if providers == nil {
		return nil, errors.New("mailbox providers are required")
	}
	return &Repository{pool: pool, providers: providers}, nil
}

func (r *Repository) recordFromRow(row *MailboxRow) *mailboxmodel.Record {
	return row.ToRecord(r.providers.NormalizeProviderInput, r.providers.PrepareProjection)
}

func safeText(value string) string {
	return redactx.TextSnippet(value, mailboxErrorSnippetLimit)
}
