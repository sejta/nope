package dbkit

import (
	"database/sql"
	stderrors "errors"

	"github.com/go-sql-driver/mysql"
)

// Kind описывает тип ошибки базы данных.
type Kind int

const (
	// KindUnknown — неизвестная ошибка.
	KindUnknown Kind = iota
	// KindNoRows — отсутствие строк.
	KindNoRows
	// KindUniqueViolation — нарушение уникальности.
	KindUniqueViolation
	// KindForeignKeyViolation — нарушение внешнего ключа.
	KindForeignKeyViolation
)

// Classify определяет тип ошибки базы данных.
func Classify(err error) Kind {
	if err == nil {
		return KindUnknown
	}
	if stderrors.Is(err, sql.ErrNoRows) {
		return KindNoRows
	}
	var my *mysql.MySQLError
	if stderrors.As(err, &my) && my != nil {
		switch my.Number {
		case 1062:
			return KindUniqueViolation
		case 1451, 1452:
			return KindForeignKeyViolation
		}
	}
	var pg interface {
		SQLState() string
	}
	if stderrors.As(err, &pg) && pg != nil {
		switch pg.SQLState() {
		case "23505":
			return KindUniqueViolation
		case "23503":
			return KindForeignKeyViolation
		}
	}
	return KindUnknown
}

// IsNoRows проверяет ошибку отсутствия строк.
func IsNoRows(err error) bool {
	return stderrors.Is(err, sql.ErrNoRows)
}

// IsUnique проверяет ошибку уникальности.
func IsUnique(err error) bool {
	return Classify(err) == KindUniqueViolation
}

// IsForeignKey проверяет ошибку внешнего ключа.
func IsForeignKey(err error) bool {
	return Classify(err) == KindForeignKeyViolation
}
