package mailboxpg

const inboxMessageInsertSQL = `
		INSERT INTO mailbox_inbox_messages (
			provider, mailbox_email, message_key, message_id, subject, from_address,
			body_preview, body_text, html_body, raw_size, received_at, recipients_json,
			source_mailbox_email,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $14)
		ON CONFLICT (provider, mailbox_email, message_key) DO UPDATE SET
			message_id = EXCLUDED.message_id,
			subject = EXCLUDED.subject,
			from_address = EXCLUDED.from_address,
			body_preview = EXCLUDED.body_preview,
			body_text = EXCLUDED.body_text,
			html_body = EXCLUDED.html_body,
			raw_size = EXCLUDED.raw_size,
			received_at = EXCLUDED.received_at,
			recipients_json = EXCLUDED.recipients_json,
			source_mailbox_email = EXCLUDED.source_mailbox_email,
			updated_at = EXCLUDED.updated_at
	`
