package eventoutbox

import (
	"context"
	"fmt"
	"time"
)

func ClaimPendingPgx(ctx context.Context, tx PgxTx, table string, batch int, now int64) ([]Row, error) {
	if tx == nil {
		return nil, ErrNilDB
	}
	tableName, err := postgresIdentifier(table)
	if err != nil {
		return nil, err
	}
	if batch <= 0 {
		batch = DefaultBatch
	}
	if now <= 0 {
		now = time.Now().Unix()
	}
	rows, err := tx.Query(ctx, fmt.Sprintf(claimPendingSQL, tableName), StatusPending, now, batch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Row{}
	for rows.Next() {
		var row Row
		if err := rows.Scan(&row.EventID, &row.Envelope, &row.AttemptCount); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
