package eventoutbox

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInvalidTableName = errors.New("event outbox table name is invalid")
	ErrNilDB            = errors.New("event outbox database handle is nil")
)

type PgxTx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func PostgresSchemaStatements(table string, pendingIndex string) ([]string, error) {
	tableName, err := postgresIdentifier(table)
	if err != nil {
		return nil, err
	}
	indexName, err := postgresIdentifier(pendingIndex)
	if err != nil {
		return nil, err
	}
	return []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			event_id TEXT PRIMARY KEY,
			subject TEXT NOT NULL,
			event_name TEXT NOT NULL,
			idempotency_key TEXT NOT NULL DEFAULT '',
			envelope BYTEA NOT NULL,
			status TEXT NOT NULL DEFAULT 'PENDING',
			attempt_count INT NOT NULL DEFAULT 0,
			next_attempt_at BIGINT NOT NULL DEFAULT 0,
			last_error TEXT NOT NULL DEFAULT '',
			published_at BIGINT NOT NULL DEFAULT 0,
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL
		)`, tableName),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s (status, next_attempt_at, created_at)`, indexName, tableName),
	}, nil
}

func postgresIdentifier(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrInvalidTableName
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return "", fmt.Errorf("%w: %s", ErrInvalidTableName, value)
	}
	return value, nil
}
