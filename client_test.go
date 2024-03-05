package dbmigrate

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"regexp"
	"testing"
)

// mockMigrationFunc is a mock migration function for testing purposes.
func mockMigrationFunc(db *gorm.DB) error {
	// This mock function simply attempts to create a dummy table.
	return db.Exec("CREATE TABLE IF NOT EXISTS dummy_table (id INT)").Error
}

// mockFailingMigrationFunc is a mock migration function that always fails.
func mockFailingMigrationFunc(db *gorm.DB) error {
	return errors.New("mock migration failure")
}

func setupMockDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: mockDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}

func TestPerformMigration(t *testing.T) {
	// Mock database setup
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm database: %v", err)
	}

	client := NewClient(gormDB)

	// Define expectations
	// Expect a query to check if the migration has already been executed
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "migrations" WHERE name = $1 ORDER BY "migrations"."id" LIMIT 1`)).
		WithArgs("test_migration").
		WillReturnRows(sqlmock.NewRows(nil)) // indicating the migration hasn't been performed

	// Adapt this to match the SQL executed by your migration function
	// For demonstration, let's expect a CREATE TABLE command
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS dummy_table \\(id INT\\)").WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectCommit()

	// Expect an insert into the migrations table
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "migrations"`)).
		WithArgs("test_migration", "test_user", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Test the PerformMigration method
	err = client.PerformMigration("test_migration", "test_user", func(db *gorm.DB) error {
		// This should match the actual migration function you're testing
		return db.Exec("CREATE TABLE IF NOT EXISTS dummy_table (id INT)").Error
	})
	if err != nil {
		t.Errorf("PerformMigration failed: %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}

func TestPerformMigrationFailure(t *testing.T) {
	db, _, err := setupMockDB()
	if err != nil {
		t.Fatalf("Failed to set up mock DB: %v", err)
	}

	client := NewClient(db)

	err = client.PerformMigration("fail_migration", "test_user", mockFailingMigrationFunc)
	if err == nil {
		t.Errorf("Expected an error, but got none")
	}
}
