package eventoutbox

import (
	"context"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/eventbus"
)

type PgxBeginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

type PgxProcessor struct {
	Beginner       PgxBeginner
	Table          string
	Publisher      eventbus.Publisher
	PublishOptions PublishOptions
}

func NewPgxProcessor(beginner PgxBeginner, table string, publisher eventbus.Publisher) *PgxProcessor {
	return &PgxProcessor{Beginner: beginner, Table: table, Publisher: publisher}
}

func (p *PgxProcessor) PublishPending(ctx context.Context, batch int) (int, error) {
	if p == nil || p.Beginner == nil || p.Publisher == nil {
		return 0, nil
	}
	if batch <= 0 {
		batch = DefaultBatch
	}
	tx, err := p.Beginner.Begin(ctx)
	if err != nil {
		return 0, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	published, err := p.publishPendingInTx(ctx, tx, batch)
	if err != nil {
		return published, err
	}
	if err := tx.Commit(ctx); err != nil {
		return published, err
	}
	committed = true
	return published, nil
}

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
