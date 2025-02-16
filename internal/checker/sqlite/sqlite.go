package sqlite

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/robalyx/rotten/internal/common"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

var ErrInvalidSchema = errors.New("invalid database schema")

// Checker implements the common.Checker interface for SQLite storage.
type Checker struct {
	dir string
}

// New creates a new SQLite checker.
func New(dir string) *Checker {
	return &Checker{dir: dir}
}

// Check verifies if the given ID exists in the SQLite database.
func (c *Checker) Check(checkType common.CheckType, id string) (*common.CheckResult, error) {
	// Determine filename based on check type
	filename := "users.db"
	tableName := "users"
	if checkType == common.CheckTypeGroup {
		filename = "groups.db"
		tableName = "groups"
	}

	// Open database
	dbPath := filepath.Join(c.dir, filename)
	conn, err := sqlite.OpenConn(dbPath, sqlite.OpenReadOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer conn.Close()

	// Validate schema
	if err := validateSchema(conn, tableName); err != nil {
		return nil, err
	}

	var result common.CheckResult
	query := fmt.Sprintf("SELECT status, reason, confidence FROM %s WHERE hash = ?", tableName)
	err = sqlitex.Execute(conn, query,
		&sqlitex.ExecOptions{
			Args: []interface{}{id},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				result = common.CheckResult{
					Found:      true,
					Status:     stmt.ColumnText(0),
					Reason:     stmt.ColumnText(1),
					Confidence: stmt.ColumnFloat(2),
				}
				return nil
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	return &result, nil
}

// GetHashCount returns the number of hashes in the database.
func (c *Checker) GetHashCount(checkType common.CheckType) (uint64, error) {
	// Determine filename based on check type
	filename := "users.db"
	tableName := "users"
	if checkType == common.CheckTypeGroup {
		filename = "groups.db"
		tableName = "groups"
	}

	// Open database
	dbPath := filepath.Join(c.dir, filename)
	conn, err := sqlite.OpenConn(dbPath, sqlite.OpenReadOnly)
	if err != nil {
		return 0, fmt.Errorf("failed to open database: %w", err)
	}
	defer conn.Close()

	// Validate schema
	if err := validateSchema(conn, tableName); err != nil {
		return 0, err
	}

	var count uint64
	query := "SELECT COUNT(*) FROM " + tableName
	err = sqlitex.Execute(conn, query,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				count = uint64(stmt.ColumnInt64(0)) //nolint:gosec
				return nil
			},
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to count hashes: %w", err)
	}

	return count, nil
}

// validateSchema checks if the table has the required columns.
func validateSchema(conn *sqlite.Conn, tableName string) error {
	err := sqlitex.Execute(conn, "SELECT hash, status, reason, confidence FROM "+tableName+" LIMIT 0",
		&sqlitex.ExecOptions{},
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSchema, err)
	}
	return nil
}
