package dbkit

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueryAllEmpty(t *testing.T) {
	db, mock := newMockDB(t)
	rows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery("SELECT id FROM users").WillReturnRows(rows)

	got, err := QueryAll(context.Background(), db, "SELECT id FROM users", nil, func(r *sql.Rows) (int, error) {
		var id int
		return id, r.Scan(&id)
	})
	if err != nil {
		t.Fatalf("ожидали nil error, получили %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("ожидали пустой слайс, получили %v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestQueryAllTwoRows(t *testing.T) {
	db, mock := newMockDB(t)
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2)
	mock.ExpectQuery("SELECT id FROM users").WillReturnRows(rows)

	got, err := QueryAll(context.Background(), db, "SELECT id FROM users", nil, func(r *sql.Rows) (int, error) {
		var id int
		return id, r.Scan(&id)
	})
	if err != nil {
		t.Fatalf("ожидали nil error, получили %v", err)
	}
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("ожидали [1 2], получили %v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestQueryAllScanError(t *testing.T) {
	db, mock := newMockDB(t)
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2)
	mock.ExpectQuery("SELECT id FROM users").WillReturnRows(rows)

	errScan := errors.New("scan error")
	calls := 0
	_, err := QueryAll(context.Background(), db, "SELECT id FROM users", nil, func(r *sql.Rows) (int, error) {
		calls++
		if calls == 2 {
			return 0, errScan
		}
		var id int
		return id, r.Scan(&id)
	})
	if !errors.Is(err, errScan) {
		t.Fatalf("ожидали scan error, получили %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestQueryOneNoRows(t *testing.T) {
	db, mock := newMockDB(t)
	rows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery("SELECT id FROM users WHERE id=?").WithArgs(1).WillReturnRows(rows)

	_, err := QueryOne(context.Background(), db, "SELECT id FROM users WHERE id=?", []any{1}, func(r *sql.Rows) (int, error) {
		var id int
		return id, r.Scan(&id)
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("ожидали sql.ErrNoRows, получили %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestQueryOneOK(t *testing.T) {
	db, mock := newMockDB(t)
	rows := sqlmock.NewRows([]string{"id"}).AddRow(10)
	mock.ExpectQuery("SELECT id FROM users WHERE id=?").WithArgs(10).WillReturnRows(rows)

	got, err := QueryOne(context.Background(), db, "SELECT id FROM users WHERE id=?", []any{10}, func(r *sql.Rows) (int, error) {
		var id int
		return id, r.Scan(&id)
	})
	if err != nil {
		t.Fatalf("ожидали nil error, получили %v", err)
	}
	if got != 10 {
		t.Fatalf("ожидали 10, получили %v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestQueryOneTooManyRows(t *testing.T) {
	db, mock := newMockDB(t)
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2)
	mock.ExpectQuery("SELECT id FROM users WHERE id=?").WithArgs(1).WillReturnRows(rows)

	_, err := QueryOne(context.Background(), db, "SELECT id FROM users WHERE id=?", []any{1}, func(r *sql.Rows) (int, error) {
		var id int
		return id, r.Scan(&id)
	})
	if !errors.Is(err, ErrTooManyRows) {
		t.Fatalf("ожидали ErrTooManyRows, получили %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestQueryOneRowsErr(t *testing.T) {
	db, mock := newMockDB(t)
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	errRows := errors.New("rows error")
	rows.RowError(0, errRows)
	mock.ExpectQuery("SELECT id FROM users WHERE id=?").WithArgs(1).WillReturnRows(rows)

	_, err := QueryOne(context.Background(), db, "SELECT id FROM users WHERE id=?", []any{1}, func(r *sql.Rows) (int, error) {
		var id int
		return id, r.Scan(&id)
	})
	if !errors.Is(err, errRows) {
		t.Fatalf("ожидали rows error, получили %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestExistsEmpty(t *testing.T) {
	db, mock := newMockDB(t)
	rows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery("SELECT id FROM users WHERE active=1").WillReturnRows(rows)

	got, err := Exists(context.Background(), db, "SELECT id FROM users WHERE active=1", nil)
	if err != nil {
		t.Fatalf("ожидали nil error, получили %v", err)
	}
	if got {
		t.Fatalf("ожидали false, получили true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}

func TestExistsOneRow(t *testing.T) {
	db, mock := newMockDB(t)
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT id FROM users WHERE active=1").WillReturnRows(rows)

	got, err := Exists(context.Background(), db, "SELECT id FROM users WHERE active=1", nil)
	if err != nil {
		t.Fatalf("ожидали nil error, получили %v", err)
	}
	if !got {
		t.Fatalf("ожидали true, получили false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ожидания sqlmock не выполнены: %v", err)
	}
}
