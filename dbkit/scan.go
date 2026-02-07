package dbkit

import (
	"context"
	"database/sql"
	"errors"
)

// Scanner описывает минимальный контракт для чтения строки.
type Scanner interface {
	Scan(dest ...any) error
}

// ErrTooManyRows возвращается, когда запрос вернул больше одной строки.
var ErrTooManyRows = errors.New("dbkit: too many rows")

// QueryAll выполняет запрос и собирает все строки в слайс.
func QueryAll[T any](
	ctx context.Context,
	c Conn,
	query string,
	args []any,
	scan func(r *sql.Rows) (T, error),
) ([]T, error) {
	rows, err := c.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []T
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// QueryOne ожидает ровно одну строку, иначе возвращает sql.ErrNoRows или ErrTooManyRows.
func QueryOne[T any](
	ctx context.Context,
	c Conn,
	query string,
	args []any,
	scan func(r *sql.Rows) (T, error),
) (T, error) {
	var zero T

	rows, err := c.QueryContext(ctx, query, args...)
	if err != nil {
		return zero, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return zero, err
		}
		return zero, sql.ErrNoRows
	}

	item, err := scan(rows)
	if err != nil {
		return zero, err
	}

	if rows.Next() {
		return zero, ErrTooManyRows
	}
	if err := rows.Err(); err != nil {
		return zero, err
	}
	return item, nil
}

// Exists проверяет наличие хотя бы одной строки по запросу.
func Exists(ctx context.Context, c Conn, query string, args []any) (bool, error) {
	rows, err := c.QueryContext(ctx, query, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	exists := rows.Next()
	if err := rows.Err(); err != nil {
		return false, err
	}
	return exists, nil
}
