package dbkit

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestExecAffected(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectExec("UPDATE users SET active=0").WillReturnResult(sqlmock.NewResult(0, 5))

	affected, err := ExecAffected(context.Background(), db, "UPDATE users SET active=0", nil)
	if err != nil {
		t.Fatalf("ожидали nil error, получили %v", err)
	}
	if affected != 5 {
		t.Fatalf("ожидали 5, получили %v", affected)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestExecOneOK(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectExec("DELETE FROM users WHERE id=?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

	err := ExecOne(context.Background(), db, "DELETE FROM users WHERE id=?", []any{1})
	if err != nil {
		t.Fatalf("ожидали nil error, получили %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestExecOneZero(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectExec("DELETE FROM users WHERE id=?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))

	err := ExecOne(context.Background(), db, "DELETE FROM users WHERE id=?", []any{1})
	var rae RowsAffectedError
	if !errors.As(err, &rae) {
		t.Fatalf("ожидали RowsAffectedError, получили %v", err)
	}
	if rae.Expected != 1 || rae.Actual != 0 {
		t.Fatalf("ожидали Expected=1 Actual=0, получили %+v", rae)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestExecOneTwo(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectExec("DELETE FROM users WHERE id=?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 2))

	err := ExecOne(context.Background(), db, "DELETE FROM users WHERE id=?", []any{1})
	var rae RowsAffectedError
	if !errors.As(err, &rae) {
		t.Fatalf("ожидали RowsAffectedError, получили %v", err)
	}
	if rae.Expected != 1 || rae.Actual != 2 {
		t.Fatalf("ожидали Expected=1 Actual=2, получили %+v", rae)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestExecAtMostOneZero(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectExec("UPDATE users SET active=1 WHERE id=?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))

	ok, err := ExecAtMostOne(context.Background(), db, "UPDATE users SET active=1 WHERE id=?", []any{1})
	if err != nil {
		t.Fatalf("ожидали nil error, получили %v", err)
	}
	if ok {
		t.Fatalf("ожидали false, получили true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestExecAtMostOneOne(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectExec("UPDATE users SET active=1 WHERE id=?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

	ok, err := ExecAtMostOne(context.Background(), db, "UPDATE users SET active=1 WHERE id=?", []any{1})
	if err != nil {
		t.Fatalf("ожидали nil error, получили %v", err)
	}
	if !ok {
		t.Fatalf("ожидали true, получили false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestExecAtMostOneTooMany(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectExec("UPDATE users SET active=1 WHERE id=?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 2))

	ok, err := ExecAtMostOne(context.Background(), db, "UPDATE users SET active=1 WHERE id=?", []any{1})
	if ok {
		t.Fatalf("ожидали false, получили true")
	}
	var rae RowsAffectedError
	if !errors.As(err, &rae) {
		t.Fatalf("ожидали RowsAffectedError, получили %v", err)
	}
	if rae.Expected != 1 || rae.Actual != 2 {
		t.Fatalf("ожидали Expected=1 Actual=2, получили %+v", rae)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}
