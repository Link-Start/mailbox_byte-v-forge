package mailboxpg

import (
	"context"
	"strings"

	"mailboxapi/internal/eventoutbox"
)

func (r *Repository) EnsureSchema(ctx context.Context, outboxTable string) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS mailboxes (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			provider TEXT NOT NULL DEFAULT '` + r.providers.DefaultKey() + `',
			created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
			updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
			last_inbox_received_at_ns BIGINT NOT NULL DEFAULT 0
		)`,
		`ALTER TABLE mailboxes ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT '` + r.providers.DefaultKey() + `'`,
		`ALTER TABLE mailboxes ADD COLUMN IF NOT EXISTS created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT`,
		`ALTER TABLE mailboxes ADD COLUMN IF NOT EXISTS updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT`,
		`ALTER TABLE mailboxes ADD COLUMN IF NOT EXISTS last_inbox_received_at_ns BIGINT NOT NULL DEFAULT 0`,
		`CREATE TABLE IF NOT EXISTS mailbox_inbox_seen (
			provider TEXT NOT NULL DEFAULT '` + r.providers.DefaultKey() + `',
			mailbox_email TEXT NOT NULL,
			message_key TEXT NOT NULL,
			seen_at BIGINT NOT NULL,
			PRIMARY KEY (provider, mailbox_email, message_key)
		)`,
		`CREATE TABLE IF NOT EXISTS mailbox_inbox_messages (
			provider TEXT NOT NULL DEFAULT '` + r.providers.DefaultKey() + `',
			mailbox_email TEXT NOT NULL,
			message_key TEXT NOT NULL,
			message_id TEXT NOT NULL DEFAULT '',
			subject TEXT NOT NULL DEFAULT '',
			from_address TEXT NOT NULL DEFAULT '',
			body_preview TEXT NOT NULL DEFAULT '',
			body_text TEXT NOT NULL DEFAULT '',
			html_body TEXT NOT NULL DEFAULT '',
			raw_size BIGINT NOT NULL DEFAULT 0,
			received_at BIGINT NOT NULL DEFAULT 0,
			recipients_json TEXT NOT NULL DEFAULT '[]',
			source_mailbox_email TEXT NOT NULL DEFAULT '',
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL,
			PRIMARY KEY (provider, mailbox_email, message_key)
		)`,
		`ALTER TABLE mailbox_inbox_seen ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT '` + r.providers.DefaultKey() + `'`,
		`ALTER TABLE mailbox_inbox_messages ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT '` + r.providers.DefaultKey() + `'`,
		`ALTER TABLE mailbox_inbox_messages ADD COLUMN IF NOT EXISTS body_text TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE mailbox_inbox_messages ADD COLUMN IF NOT EXISTS html_body TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE mailbox_inbox_messages ADD COLUMN IF NOT EXISTS raw_size BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE mailbox_inbox_messages ADD COLUMN IF NOT EXISTS source_mailbox_email TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE mailbox_inbox_messages DROP COLUMN IF EXISTS otp`,
		`ALTER TABLE mailbox_inbox_messages DROP COLUMN IF EXISTS event_type`,
		`DROP TABLE IF EXISTS mailbox_latest_otps`,
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'mailbox_inbox_seen_pkey'
				  AND conrelid = 'mailbox_inbox_seen'::regclass
			) THEN
				ALTER TABLE mailbox_inbox_seen DROP CONSTRAINT mailbox_inbox_seen_pkey;
			END IF;
		END $$`,
		`ALTER TABLE mailbox_inbox_seen ADD PRIMARY KEY (provider, mailbox_email, message_key)`,
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'mailbox_inbox_messages_pkey'
				  AND conrelid = 'mailbox_inbox_messages'::regclass
			) THEN
				ALTER TABLE mailbox_inbox_messages DROP CONSTRAINT mailbox_inbox_messages_pkey;
			END IF;
		END $$`,
		`ALTER TABLE mailbox_inbox_messages ADD PRIMARY KEY (provider, mailbox_email, message_key)`,
		`DROP INDEX IF EXISTS idx_mailboxes_assigned_account`,
		`ALTER TABLE mailboxes DROP COLUMN IF EXISTS assigned_account_id`,
		`DROP INDEX IF EXISTS idx_mailboxes_status`,
		`DROP INDEX IF EXISTS idx_mailboxes_primary`,
		`DROP INDEX IF EXISTS idx_mailboxes_auth_status`,
		`ALTER TABLE mailboxes DROP COLUMN IF EXISTS status`,
		`ALTER TABLE mailboxes DROP COLUMN IF EXISTS is_primary`,
		`ALTER TABLE mailboxes DROP COLUMN IF EXISTS primary_email`,
		`ALTER TABLE mailboxes DROP COLUMN IF EXISTS password`,
		`ALTER TABLE mailboxes DROP COLUMN IF EXISTS refresh_token`,
		`ALTER TABLE mailboxes DROP COLUMN IF EXISTS access_token`,
		`ALTER TABLE mailboxes DROP COLUMN IF EXISTS auth_status`,
		`ALTER TABLE mailboxes DROP COLUMN IF EXISTS last_error`,
		`CREATE INDEX IF NOT EXISTS idx_mailboxes_provider ON mailboxes(provider)`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_inbox_seen_at ON mailbox_inbox_seen(seen_at)`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_inbox_messages_received_at ON mailbox_inbox_messages(mailbox_email, received_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_inbox_messages_provider_received_at ON mailbox_inbox_messages(provider, mailbox_email, received_at DESC)`,
	}
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

func insertSchemaStatementsAfter(statements []string, prefix string, additions []string) []string {
	if len(additions) == 0 {
		return statements
	}
	for i, statement := range statements {
		if strings.HasPrefix(strings.TrimSpace(statement), prefix) {
			out := make([]string, 0, len(statements)+len(additions))
			out = append(out, statements[:i+1]...)
			out = append(out, additions...)
			out = append(out, statements[i+1:]...)
			return out
		}
	}
	return append(statements, additions...)
}
