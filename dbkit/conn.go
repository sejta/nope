package dbkit

import (
	"context"
	"database/sql"
)

// Conn описывает минимальный набор контекстных методов для работы с БД.
type Conn interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
