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
