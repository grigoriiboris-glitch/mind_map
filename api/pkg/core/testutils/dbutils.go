package testutils

import (
	"os"
	"testing"
	"time"
	"go.etcd.io/bbolt"
)

// SetupTestDB создает временную базу данных для тестов
func SetupTestBBoltDB(t *testing.T) *bbolt.DB {
	t.Helper()
	
	// Создаем временный файл для базы данных
	tmpfile, err := os.CreateTemp("", "testdb-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()

	db, err := bbolt.Open(tmpfile.Name(), 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpfile.Name()) // Удаляем временный файл после теста
	})

	return db
}