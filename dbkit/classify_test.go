package dbkit

import (
	"database/sql"
	stderrors "errors"
	"testing"

	"github.com/go-sql-driver/mysql"
)

type pgErr struct {
	code string
}

func (e pgErr) Error() string {
	return "pg"
}

func (e pgErr) SQLState() string {
	return e.code
}

func TestClassifyMySQL(t *testing.T) {
	if got := Classify(&mysql.MySQLError{Number: 1062}); got != KindUniqueViolation {
		t.Fatalf("ожидали KindUniqueViolation, получили %v", got)
	}
	if got := Classify(&mysql.MySQLError{Number: 1451}); got != KindForeignKeyViolation {
		t.Fatalf("ожидали KindForeignKeyViolation, получили %v", got)
	}
	if got := Classify(&mysql.MySQLError{Number: 1452}); got != KindForeignKeyViolation {
		t.Fatalf("ожидали KindForeignKeyViolation, получили %v", got)
	}
	if got := Classify(&mysql.MySQLError{Number: 1}); got != KindUnknown {
		t.Fatalf("ожидали KindUnknown, получили %v", got)
	}
}

func TestClassifyNoRows(t *testing.T) {
	if got := Classify(sql.ErrNoRows); got != KindNoRows {
		t.Fatalf("ожидали KindNoRows, получили %v", got)
	}
	if !IsNoRows(sql.ErrNoRows) {
		t.Fatalf("ожидали IsNoRows=true")
	}
}

func TestClassifyWrapped(t *testing.T) {
	err := stderrors.Join(&mysql.MySQLError{Number: 1062})
	if got := Classify(err); got != KindUniqueViolation {
		t.Fatalf("ожидали KindUniqueViolation, получили %v", got)
	}
}

func TestClassifyPostgresBestEffort(t *testing.T) {
	if got := Classify(pgErr{code: "23505"}); got != KindUniqueViolation {
		t.Fatalf("ожидали KindUniqueViolation, получили %v", got)
	}
	if got := Classify(pgErr{code: "23503"}); got != KindForeignKeyViolation {
		t.Fatalf("ожидали KindForeignKeyViolation, получили %v", got)
	}
}

func TestClassifyUnknown(t *testing.T) {
	if got := Classify(stderrors.New("x")); got != KindUnknown {
		t.Fatalf("ожидали KindUnknown, получили %v", got)
	}
}
