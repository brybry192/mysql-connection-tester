package main

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestFetchSeedValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "mysql")

	// Expect the seed query using a regular expression
	mock.ExpectQuery("SELECT id FROM users ORDER BY RAND\\(\\) LIMIT 5").WillReturnRows(
		sqlmock.NewRows([]string{"id"}).
			AddRow(1).
			AddRow(2).
			AddRow(3),
	)

	results, err := fetchSeedValues(sqlxDB, "SELECT id FROM users ORDER BY RAND() LIMIT 5")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedResult := [][]interface{}{
		{int64(1)},
		{int64(2)},
		{int64(3)},
	}

	if len(results) != len(expectedResult) {
		t.Fatalf("Expected %d results, got %d", len(expectedResult), len(results))
	}

	for i, result := range results {
		if result[0] != expectedResult[i][0] {
			t.Errorf("Expected %v, got %v", expectedResult[i], result)
		}
	}
}

// Test the executeQueryWithValues function
func TestExecuteQueryWithValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "mysql")

	// Expect the query with the provided value
	mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\?").WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user", "name"}).AddRow(1, "user1", "Name One"))

	err = executeQueryWithValues(sqlxDB, "SELECT * FROM users WHERE id = ?", []interface{}{1})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}
