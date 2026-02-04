package dbkit

import (
	"context"
	"database/sql"
)

// InTx выполняет функцию в транзакции и управляет commit/rollback.
func InTx(ctx context.Context, db *sql.DB, fn func(ctx context.Context, tx Conn) error) error {
	if db == nil {
		return sql.ErrConnDone
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if rec := recover(); rec != nil {
			_ = rollbackIgnoreDone(tx)
			panic(rec)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		_ = rollbackIgnoreDone(tx)
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func rollbackIgnoreDone(tx *sql.Tx) error {
	err := tx.Rollback()
	if err == sql.ErrTxDone {
		return nil
	}
	return err
}
