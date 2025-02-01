package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/robalyx/rotten/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func setupTestFiles(t *testing.T, dir string) {
	files := map[string]string{
		filepath.Join(dir, "users.db"):  "users",
		filepath.Join(dir, "groups.db"): "groups",
	}

	for file, tableName := range files {
		conn, err := sqlite.OpenConn(file, sqlite.OpenCreate|sqlite.OpenReadWrite)
		require.NoError(t, err)
		defer conn.Close()

		// Create table with required schema
		err = sqlitex.ExecScript(conn, `
			CREATE TABLE IF NOT EXISTS `+tableName+` (
				hash TEXT PRIMARY KEY,
				status TEXT NOT NULL,
				reason TEXT NOT NULL
			);
		`)
		require.NoError(t, err)

		// Insert test data for users.db
		if tableName == "users" {
			err = sqlitex.ExecScript(conn, `
				INSERT INTO users (hash, status, reason)
				VALUES ('testHash123', 'banned', 'violation');
			`)
			require.NoError(t, err)
		}
	}
}

func TestNew(t *testing.T) {
	dir := "test_dir"
	checker := New(dir)
	assert.NotNil(t, checker)
	assert.Equal(t, dir, checker.dir)
}

func TestChecker_Check(t *testing.T) {
	tempDir := t.TempDir()
	setupTestFiles(t, tempDir)

	tests := []struct {
		name        string
		checkType   common.CheckType
		hash        string
		wantFound   bool
		wantStatus  string
		wantReason  string
		wantErrType error
	}{
		{
			name:      "Valid user check - not found",
			checkType: common.CheckTypeUser,
			hash:      "0123456789abcdef",
			wantFound: false,
		},
		{
			name:      "Valid group check - not found",
			checkType: common.CheckTypeGroup,
			hash:      "0123456789abcdef",
			wantFound: false,
		},
		{
			name:      "Valid friends check - not found",
			checkType: common.CheckTypeFriends,
			hash:      "0123456789abcdef",
			wantFound: false,
		},
		{
			name:       "Valid user check - found",
			checkType:  common.CheckTypeUser,
			hash:       "testHash123",
			wantFound:  true,
			wantStatus: "banned",
			wantReason: "violation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := New(tempDir)
			found, status, reason, err := checker.Check(tt.checkType, tt.hash)

			if tt.wantErrType != nil {
				assert.ErrorIs(t, err, tt.wantErrType)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantFound, found)
			if tt.wantFound {
				assert.Equal(t, tt.wantStatus, status)
				assert.Equal(t, tt.wantReason, reason)
			} else {
				assert.Empty(t, status)
				assert.Empty(t, reason)
			}
		})
	}
}

func TestChecker_GetHashCount(t *testing.T) {
	tempDir := t.TempDir()
	setupTestFiles(t, tempDir)

	tests := []struct {
		name        string
		checkType   common.CheckType
		wantCount   uint64
		wantErrType error
	}{
		{
			name:      "Valid user check",
			checkType: common.CheckTypeUser,
			wantCount: 1, // One record in test file
		},
		{
			name:      "Valid group check",
			checkType: common.CheckTypeGroup,
			wantCount: 0, // Empty test file
		},
		{
			name:      "Valid friends check",
			checkType: common.CheckTypeFriends,
			wantCount: 1, // One record in test file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := New(tempDir)
			count, err := checker.GetHashCount(tt.checkType)

			if tt.wantErrType != nil {
				assert.ErrorIs(t, err, tt.wantErrType)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestChecker_NonexistentFile(t *testing.T) {
	tempDir := t.TempDir()
	checker := New(tempDir)

	// Test Check with nonexistent file
	found, status, reason, err := checker.Check(common.CheckTypeUser, "0123456789abcdef")
	assert.Error(t, err)
	assert.False(t, found)
	assert.Empty(t, status)
	assert.Empty(t, reason)

	// Test GetHashCount with nonexistent file
	count, err := checker.GetHashCount(common.CheckTypeUser)
	assert.Error(t, err)
	assert.Zero(t, count)
}

func TestChecker_InvalidDatabase(t *testing.T) {
	tempDir := t.TempDir()

	// Create file with invalid SQLite format
	err := os.WriteFile(filepath.Join(tempDir, "users.db"), []byte("invalid"), 0o600)
	require.NoError(t, err)

	checker := New(tempDir)

	// Test Check with invalid database
	found, status, reason, err := checker.Check(common.CheckTypeUser, "0123456789abcdef")
	assert.Error(t, err)
	assert.False(t, found)
	assert.Empty(t, status)
	assert.Empty(t, reason)

	// Test GetHashCount with invalid database
	count, err := checker.GetHashCount(common.CheckTypeUser)
	assert.Error(t, err)
	assert.Zero(t, count)
}

func TestChecker_InvalidSchema(t *testing.T) {
	tempDir := t.TempDir()

	// Create database with invalid schema
	dbPath := filepath.Join(tempDir, "users.db")
	conn, err := sqlite.OpenConn(dbPath, sqlite.OpenCreate|sqlite.OpenReadWrite)
	require.NoError(t, err)

	err = sqlitex.ExecScript(conn, `
		CREATE TABLE users (
			hash TEXT PRIMARY KEY,
			invalid_column TEXT
		);
	`)
	require.NoError(t, err)
	conn.Close()

	checker := New(tempDir)

	// Test Check with invalid schema
	found, status, reason, err := checker.Check(common.CheckTypeUser, "0123456789abcdef")
	assert.Error(t, err)
	assert.False(t, found)
	assert.Empty(t, status)
	assert.Empty(t, reason)

	// Test GetHashCount with invalid schema
	count, err := checker.GetHashCount(common.CheckTypeUser)
	assert.Error(t, err)
	assert.Zero(t, count)
}

func TestChecker_EmptyDatabase(t *testing.T) {
	tempDir := t.TempDir()

	// Create empty database with correct schema
	dbPath := filepath.Join(tempDir, "users.db")
	conn, err := sqlite.OpenConn(dbPath, sqlite.OpenCreate|sqlite.OpenReadWrite)
	require.NoError(t, err)

	err = sqlitex.ExecScript(conn, `
		CREATE TABLE users (
			hash TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			reason TEXT NOT NULL
		);
	`)
	require.NoError(t, err)
	conn.Close()

	checker := New(tempDir)

	// Test Check with empty database
	found, status, reason, err := checker.Check(common.CheckTypeUser, "0123456789abcdef")
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Empty(t, status)
	assert.Empty(t, reason)

	// Test GetHashCount with empty database
	count, err := checker.GetHashCount(common.CheckTypeUser)
	assert.NoError(t, err)
	assert.Zero(t, count)
}
