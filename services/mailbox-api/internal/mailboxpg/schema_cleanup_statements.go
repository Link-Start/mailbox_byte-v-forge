package mailboxpg

func legacySchemaStatements() []string {
	return []string{
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
	}
}

func indexSchemaStatements() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_mailboxes_provider ON mailboxes(provider)`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_inbox_seen_at ON mailbox_inbox_seen(seen_at)`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_inbox_messages_received_at ON mailbox_inbox_messages(mailbox_email, received_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_inbox_messages_provider_received_at ON mailbox_inbox_messages(provider, mailbox_email, received_at DESC)`,
	}
}
