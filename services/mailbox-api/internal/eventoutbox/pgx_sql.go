package eventoutbox

const insertRecordSQL = `
		INSERT INTO %s (
			event_id, subject, event_name, idempotency_key, envelope, status,
			attempt_count, next_attempt_at, last_error, published_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, 0, 0, '', 0, $7, $7)
		ON CONFLICT (event_id) DO NOTHING
	`

const claimPendingSQL = `
		SELECT event_id, envelope, attempt_count
		FROM %s
		WHERE status = $1 AND next_attempt_at <= $2
		ORDER BY created_at ASC
		LIMIT $3
		FOR UPDATE SKIP LOCKED
	`
