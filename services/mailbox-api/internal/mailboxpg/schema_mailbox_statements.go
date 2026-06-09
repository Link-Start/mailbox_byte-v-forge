package mailboxpg

func mailboxSchemaStatements(defaultProvider string) []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS mailboxes (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			provider TEXT NOT NULL DEFAULT '` + defaultProvider + `',
			created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
			updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
			last_inbox_received_at_ns BIGINT NOT NULL DEFAULT 0
		)`,
		`ALTER TABLE mailboxes ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT '` + defaultProvider + `'`,
		`ALTER TABLE mailboxes ADD COLUMN IF NOT EXISTS created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT`,
		`ALTER TABLE mailboxes ADD COLUMN IF NOT EXISTS updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT`,
		`ALTER TABLE mailboxes ADD COLUMN IF NOT EXISTS last_inbox_received_at_ns BIGINT NOT NULL DEFAULT 0`,
	}
}
