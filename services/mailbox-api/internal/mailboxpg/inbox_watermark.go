package mailboxpg

import (
	"context"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/jackc/pgx/v5"
)

func trackInboxWatermark(watermarks map[string]int64, mailboxEmail string, receivedAtUnix int64) {
	if receivedAtUnix <= 0 {
		return
	}
	watermark := time.Unix(receivedAtUnix, 0).UnixNano()
	if watermarks[mailboxEmail] < watermark {
		watermarks[mailboxEmail] = watermark
	}
}

func UpdateInboxWatermarks(ctx context.Context, tx pgx.Tx, watermarks map[string]int64, now int64) error {
	for mailboxEmail, watermark := range watermarks {
		if watermark <= 0 {
			continue
		}
		if _, err := tx.Exec(ctx, `
			UPDATE mailboxes
			SET last_inbox_received_at_ns = GREATEST(last_inbox_received_at_ns, $1), updated_at = $2
			WHERE email = $3
		`, watermark, now, emailx.Normalize(mailboxEmail)); err != nil {
			return err
		}
	}
	return nil
}
