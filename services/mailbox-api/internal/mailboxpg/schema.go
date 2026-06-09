package mailboxpg

import (
	"context"

	"mailboxapi/internal/eventoutbox"
)

func (r *Repository) EnsureSchema(ctx context.Context, outboxTable string) error {
	statements := baseSchemaStatements(r.providers.DefaultKey())
	outboxStatements, err := eventoutbox.PostgresSchemaStatements(outboxTable, "idx_mailbox_event_outbox_pending")
	if err != nil {
		return err
	}
	statements = append(statements, outboxStatements...)
	statements = insertSchemaStatementsAfter(statements, "ALTER TABLE mailboxes ADD COLUMN IF NOT EXISTS last_inbox_received_at_ns", r.providers.SchemaStatements())
	statements = insertSchemaStatementsAfter(statements, "ALTER TABLE mailboxes DROP COLUMN IF EXISTS assigned_account_id", r.providers.LegacyStatements())
	for _, statement := range statements {
		if _, err := r.pool.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
