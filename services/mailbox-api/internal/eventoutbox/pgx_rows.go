package eventoutbox

import (
	"context"
	"fmt"
	"time"
)

func InsertRecordPgx(ctx context.Context, tx PgxTx, table string, record Record, now int64) error {
	if tx == nil {
		return ErrNilDB
	}
	tableName, err := postgresIdentifier(table)
	if err != nil {
		return err
	}
	if now <= 0 {
		now = time.Now().Unix()
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(insertRecordSQL, tableName), record.EventID, record.Subject, record.EventName, record.IdempotencyKey, record.Envelope, StatusPending, now)
	return err
}
