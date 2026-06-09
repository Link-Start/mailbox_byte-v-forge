package mailboxpg

func inboxSchemaStatements(defaultProvider string) []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS mailbox_inbox_seen (
			provider TEXT NOT NULL DEFAULT '` + defaultProvider + `',
			mailbox_email TEXT NOT NULL,
			message_key TEXT NOT NULL,
			seen_at BIGINT NOT NULL,
			PRIMARY KEY (provider, mailbox_email, message_key)
		)`,
		`CREATE TABLE IF NOT EXISTS mailbox_inbox_messages (
			provider TEXT NOT NULL DEFAULT '` + defaultProvider + `',
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
		`ALTER TABLE mailbox_inbox_seen ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT '` + defaultProvider + `'`,
		`ALTER TABLE mailbox_inbox_messages ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT '` + defaultProvider + `'`,
		`ALTER TABLE mailbox_inbox_messages ADD COLUMN IF NOT EXISTS body_text TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE mailbox_inbox_messages ADD COLUMN IF NOT EXISTS html_body TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE mailbox_inbox_messages ADD COLUMN IF NOT EXISTS raw_size BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE mailbox_inbox_messages ADD COLUMN IF NOT EXISTS source_mailbox_email TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE mailbox_inbox_messages DROP COLUMN IF EXISTS otp`,
		`ALTER TABLE mailbox_inbox_messages DROP COLUMN IF EXISTS event_type`,
		`DROP TABLE IF EXISTS mailbox_latest_otps`,
		inboxSeenPrimaryKeyStatement,
		`ALTER TABLE mailbox_inbox_seen ADD PRIMARY KEY (provider, mailbox_email, message_key)`,
		inboxMessagesPrimaryKeyStatement,
		`ALTER TABLE mailbox_inbox_messages ADD PRIMARY KEY (provider, mailbox_email, message_key)`,
	}
}

const inboxSeenPrimaryKeyStatement = `DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'mailbox_inbox_seen_pkey'
				  AND conrelid = 'mailbox_inbox_seen'::regclass
			) THEN
				ALTER TABLE mailbox_inbox_seen DROP CONSTRAINT mailbox_inbox_seen_pkey;
			END IF;
		END $$`

const inboxMessagesPrimaryKeyStatement = `DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'mailbox_inbox_messages_pkey'
				  AND conrelid = 'mailbox_inbox_messages'::regclass
			) THEN
				ALTER TABLE mailbox_inbox_messages DROP CONSTRAINT mailbox_inbox_messages_pkey;
			END IF;
		END $$`
