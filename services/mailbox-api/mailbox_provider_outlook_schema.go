package main

func outlookSchemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS mailbox_outlook_accounts (
			mailbox_email TEXT PRIMARY KEY REFERENCES mailboxes(email) ON DELETE CASCADE,
			password TEXT NOT NULL DEFAULT '',
			refresh_token TEXT NOT NULL DEFAULT '',
			access_token TEXT NOT NULL DEFAULT '',
			auth_status TEXT NOT NULL DEFAULT 'OAUTH_PENDING',
			last_error TEXT NOT NULL DEFAULT '',
			created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
			updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_outlook_accounts_auth_status ON mailbox_outlook_accounts(auth_status)`,
		`CREATE INDEX IF NOT EXISTS idx_mailbox_outlook_accounts_refresh_token ON mailbox_outlook_accounts(refresh_token)`,
	}
}

func outlookLegacyStatements() []string {
	return []string{
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'mailboxes' AND column_name = 'password'
			) THEN
				EXECUTE $sql$
					INSERT INTO mailbox_outlook_accounts (
						mailbox_email, password, refresh_token, access_token,
						auth_status, last_error, created_at, updated_at
					)
					SELECT email, password, refresh_token, access_token,
						CASE
							WHEN auth_status <> '' THEN auth_status
							WHEN refresh_token <> '' THEN 'AUTHORIZED'
							ELSE 'OAUTH_PENDING'
						END,
						last_error, created_at, updated_at
					FROM mailboxes
					WHERE LOWER(COALESCE(provider, 'outlook')) IN ('', 'outlook', 'microsoft', 'graph')
					ON CONFLICT (mailbox_email) DO UPDATE SET
						password = CASE WHEN EXCLUDED.password <> '' THEN EXCLUDED.password ELSE mailbox_outlook_accounts.password END,
						refresh_token = CASE WHEN EXCLUDED.refresh_token <> '' THEN EXCLUDED.refresh_token ELSE mailbox_outlook_accounts.refresh_token END,
						access_token = CASE WHEN EXCLUDED.access_token <> '' THEN EXCLUDED.access_token ELSE mailbox_outlook_accounts.access_token END,
						auth_status = CASE WHEN EXCLUDED.auth_status <> '' THEN EXCLUDED.auth_status ELSE mailbox_outlook_accounts.auth_status END,
						last_error = CASE WHEN EXCLUDED.last_error <> '' THEN EXCLUDED.last_error ELSE mailbox_outlook_accounts.last_error END,
						updated_at = GREATEST(mailbox_outlook_accounts.updated_at, EXCLUDED.updated_at)
				$sql$;
			END IF;
		END $$`,
		`UPDATE mailbox_outlook_accounts SET auth_status = 'OAUTH_PENDING', last_error = ''
			WHERE auth_status = 'AUTH_FAILED'
			AND last_error = 'registered mailbox has no OAuth refresh token'`,
		`DELETE FROM mailboxes alias
			USING mailboxes base
			WHERE alias.provider = 'outlook'
			  AND base.provider = alias.provider
			  AND split_part(alias.email, '@', 1) LIKE '%+%'
			  AND base.email = regexp_replace(alias.email, '^([^+@]+)\+[^@]*@(.+)$', '\1@\2')
			  AND base.email <> alias.email`,
	}
}
