package sqlvertest

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

var (
	PostgresUsername = "yourusername"
	PostgresPassword = "yourpassword"
)

type PostgresConnectionError struct {
	Err error
}

func (e *PostgresConnectionError) Error() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "\nConnection to Postgres failed: %s\n", e.Err.Error())
	fmt.Fprintln(&b, "Please run the following commands to create the necessary user and grant permissions:")
	fmt.Fprintf(&b, "\tsudo -u postgres createuser --createdb --pwprompt %s\n", PostgresUsername)
	fmt.Fprintf(&b, "\tsudo -u postgres psql -c \"ALTER USER %s WITH PASSWORD '%s' CREATEDB;\"\n", PostgresUsername, PostgresPassword)
	return b.String()
}

func (e *PostgresConnectionError) Unwrap() error {
	return e.Err
}

func generateUUID() (string, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}

	// Set the version (4) and variant (RFC4122)
	uuid[6] = (uuid[6] & 0x0F) | 0x40
	uuid[8] = (uuid[8] & 0x3F) | 0x80

	// Format the UUID as a string with underscores
	uuidStr := fmt.Sprintf("%x_%x_%x_%x_%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])

	return uuidStr, nil
}

func connectToPostgres() (*sql.DB, error) {
	connectionString := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=postgres sslmode=disable", PostgresUsername, PostgresPassword)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, &PostgresConnectionError{Err: err}
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, &PostgresConnectionError{Err: err}
	}
	return db, nil
}

// DB returns a fresh postgres database for a test to use and registers cleanup
// so that the database is dropped when the test succeeds, but kept intact
// when it fails.
func DB(t *testing.T) *sql.DB {
	t.Helper()
	// Connect to the main database to create a test database
	mainDB, err := connectToPostgres()
	if err != nil {
		t.Fatal(err)
	}
	defer mainDB.Close()

	// Generate a UUID-like name for the test database
	uuid, err := generateUUID()
	if err != nil {
		t.Fatalf("failed to generate UUID for test database name: %s", err)
	}
	testDBName := fmt.Sprintf("test_%s", uuid)

	// Create the test database
	_, err = mainDB.Exec(fmt.Sprintf("CREATE DATABASE %s;", testDBName))
	if err != nil {
		t.Fatalf("failed to create test database: %s", err)
	}

	// Connect to the test database
	testDBConnectionString := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable", PostgresUsername, PostgresPassword, testDBName)
	testDB, err := sql.Open("postgres", testDBConnectionString)
	if err != nil {
		t.Fatalf("failed to connect to test database: %s", err)
	}

	// Check the connection to the test database
	err = testDB.Ping()
	if err != nil {
		_ = testDB.Close()
		t.Fatalf("failed to ping test database: %s", err)
	}
	t.Cleanup(func() {
		if err := testDB.Close(); err != nil {
			t.Errorf("Failed to close test DB connection: %s", err)
		}
		if t.Failed() {
			t.Logf("Database left intact: %q", testDBName)
			return
		}
		mainDB, err := connectToPostgres()
		if err != nil {
			t.Fatalf("cannot connect to postgres to drop %q: %s", testDBName, err)
		}
		defer mainDB.Close()
		_, err = mainDB.Exec(fmt.Sprintf("DROP DATABASE %s;", testDBName))
		if err != nil {
			t.Fatalf("failed to drop test database %q: %s", testDBName, err)
		}
	})
	return testDB
}
