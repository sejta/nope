package dbkit

import (
	"context"
	"strconv"
)

// RowsAffectedError описывает несоответствие ожидаемого и фактического числа строк.
type RowsAffectedError struct {
	Expected int64
	Actual   int64
}

// Error возвращает текст ошибки для RowsAffectedError.
func (e RowsAffectedError) Error() string {
	return "dbkit: expected rows affected " + strconv.FormatInt(e.Expected, 10) + ", got " + strconv.FormatInt(e.Actual, 10)
}

// ExecAffected выполняет запрос и возвращает количество затронутых строк.
func ExecAffected(ctx context.Context, c Conn, query string, args []any) (int64, error) {
	res, err := c.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

// ExecOne ожидает ровно одну затронутую строку.
func ExecOne(ctx context.Context, c Conn, query string, args []any) error {
	affected, err := ExecAffected(ctx, c, query, args)
	if err != nil {
		return err
	}
	if affected == 1 {
		return nil
	}
	return RowsAffectedError{Expected: 1, Actual: affected}
}

// ExecAtMostOne допускает 0 или 1 строку и возвращает true при ровно одной.
func ExecAtMostOne(ctx context.Context, c Conn, query string, args []any) (bool, error) {
	affected, err := ExecAffected(ctx, c, query, args)
	if err != nil {
		return false, err
	}
	switch affected {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, RowsAffectedError{Expected: 1, Actual: affected}
	}
}
