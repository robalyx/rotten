package checker

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
	// Create and initialize SQLite test files
	sqliteFiles := map[string]string{
		filepath.Join(dir, "users.db"):  "users",
		filepath.Join(dir, "groups.db"): "groups",
	}
	for file, tableName := range sqliteFiles {
		conn, err := sqlite.OpenConn(file, sqlite.OpenCreate|sqlite.OpenReadWrite)
		require.NoError(t, err)
		defer conn.Close()

		// Create table with required schema
		err = sqlitex.ExecScript(conn, `
			CREATE TABLE IF NOT EXISTS `+tableName+` (
				hash TEXT PRIMARY KEY,
				status TEXT NOT NULL,
				reason TEXT NOT NULL,
				confidence REAL NOT NULL DEFAULT 1.0
			);
		`)
		require.NoError(t, err)
	}

	// Create binary test files
	binaryFiles := []string{
		filepath.Join(dir, "users.bin"),
		filepath.Join(dir, "groups.bin"),
	}
	for _, file := range binaryFiles {
		f, err := os.Create(file)
		require.NoError(t, err)
		// Write a simple header (count = 0)
		_, err = f.Write([]byte{0, 0, 0, 0})
		require.NoError(t, err)
		f.Close()
	}

	// Create CSV test files
	csvFiles := []string{
		filepath.Join(dir, "users.csv"),
		filepath.Join(dir, "groups.csv"),
	}
	for _, file := range csvFiles {
		f, err := os.Create(file)
		require.NoError(t, err)
		// Write CSV header
		_, err = f.WriteString("hash,status,reason,confidence\n")
		require.NoError(t, err)
		f.Close()
	}
}

func TestNew(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		storageType common.StorageType
		wantError   bool
	}{
		{
			name:        "SQLite storage",
			storageType: common.StorageTypeSQLite,
			wantError:   false,
		},
		{
			name:        "Binary storage",
			storageType: common.StorageTypeBinary,
			wantError:   false,
		},
		{
			name:        "CSV storage",
			storageType: common.StorageTypeCSV,
			wantError:   false,
		},
		{
			name:        "Invalid storage type",
			storageType: "invalid",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker, err := New(tempDir, tt.storageType)
			if tt.wantError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrUnsupportedStorageType)
				assert.Nil(t, checker)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, checker)
			}
		})
	}
}

func TestChecker_Integration(t *testing.T) {
	tempDir := t.TempDir()
	setupTestFiles(t, tempDir)

	testHash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	tests := []struct {
		name        string
		storageType common.StorageType
		checkType   common.CheckType
	}{
		{
			name:        "SQLite user check",
			storageType: common.StorageTypeSQLite,
			checkType:   common.CheckTypeUser,
		},
		{
			name:        "SQLite group check",
			storageType: common.StorageTypeSQLite,
			checkType:   common.CheckTypeGroup,
		},
		{
			name:        "Binary user check",
			storageType: common.StorageTypeBinary,
			checkType:   common.CheckTypeUser,
		},
		{
			name:        "Binary group check",
			storageType: common.StorageTypeBinary,
			checkType:   common.CheckTypeGroup,
		},
		{
			name:        "CSV user check",
			storageType: common.StorageTypeCSV,
			checkType:   common.CheckTypeUser,
		},
		{
			name:        "CSV group check",
			storageType: common.StorageTypeCSV,
			checkType:   common.CheckTypeGroup,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker, err := New(tempDir, tt.storageType)
			require.NoError(t, err)
			require.NotNil(t, checker)

			// Test GetHashCount
			count, err := checker.GetHashCount(tt.checkType)
			assert.NoError(t, err)
			assert.Zero(t, count) // Empty test files should have 0 hashes

			// Test Check with valid but non-existent hash
			result, err := checker.Check(tt.checkType, testHash)
			assert.NoError(t, err)
			assert.False(t, result.Found)
			assert.Empty(t, result.Status)
			assert.Empty(t, result.Reason)
		})
	}
}

func TestChecker_InvalidDirectory(t *testing.T) {
	nonexistentDir := filepath.Join(t.TempDir(), "nonexistent")
	testHash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	tests := []struct {
		name        string
		storageType common.StorageType
	}{
		{
			name:        "SQLite storage",
			storageType: common.StorageTypeSQLite,
		},
		{
			name:        "Binary storage",
			storageType: common.StorageTypeBinary,
		},
		{
			name:        "CSV storage",
			storageType: common.StorageTypeCSV,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker, err := New(nonexistentDir, tt.storageType)
			require.NoError(t, err) // New should not fail with invalid directory
			require.NotNil(t, checker)

			// GetHashCount should fail
			_, err = checker.GetHashCount(common.CheckTypeUser)
			assert.Error(t, err)

			// Check should fail
			result, err := checker.Check(common.CheckTypeUser, testHash)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}
