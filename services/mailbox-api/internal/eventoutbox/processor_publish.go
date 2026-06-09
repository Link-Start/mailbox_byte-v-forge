package eventoutbox

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (p *PgxProcessor) publishPendingInTx(ctx context.Context, tx pgx.Tx, batch int) (int, error) {
	rows, err := ClaimPendingPgx(ctx, tx, p.Table, batch, optionUnix(p.PublishOptions))
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	updates, err := NewPgxUpdates(tx, p.Table)
	if err != nil {
		return 0, err
	}
	return PublishRows(ctx, p.Publisher, rows, updates, p.PublishOptions)
}

func optionUnix(options PublishOptions) int64 {
	if options.Now == nil {
		return 0
	}
	return options.Now().Unix()
}
