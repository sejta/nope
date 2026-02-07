package dbkit

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать sqlmock: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db, mock
}
