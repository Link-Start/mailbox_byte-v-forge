package mailboxpg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (b *providerStorageBuilder) exec(ctx context.Context, tx pgx.Tx, extraArgs ...any) error {
	if b.err != nil {
		return b.err
	}
	updates := b.orderedUpdates()
	args := append([]any{}, b.args...)
	args = append(args, extraArgs...)
	query := b.insertSQL(updates)
	_, err := tx.Exec(ctx, query, args...)
	return err
}

func (b *providerStorageBuilder) insertSQL(updates []string) string {
	placeholders := b.placeholders()
	if len(updates) == 0 {
		return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO NOTHING",
			b.table,
			strings.Join(b.columns, ", "),
			strings.Join(placeholders, ", "),
			b.emailColumn,
		)
	}
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		b.table,
		strings.Join(b.columns, ", "),
		strings.Join(placeholders, ", "),
		b.emailColumn,
		strings.Join(updates, ", "),
	)
}

func (b *providerStorageBuilder) orderedUpdates() []string {
	updates := make([]string, 0, len(b.updates))
	for _, column := range b.columns {
		if update := b.updates[column]; update != "" {
			updates = append(updates, update)
		}
	}
	return updates
}
