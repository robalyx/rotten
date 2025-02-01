package checker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/robalyx/rotten/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	require.NotNil(t, v)
	require.NotNil(t, v.requiredFiles)

	// Verify required files are properly initialized
	assert.Equal(t, "users.db", v.requiredFiles[common.CheckTypeUser][common.StorageTypeSQLite])
	assert.Equal(t, "users.bin", v.requiredFiles[common.CheckTypeUser][common.StorageTypeBinary])
	assert.Equal(t, "users.csv", v.requiredFiles[common.CheckTypeUser][common.StorageTypeCSV])
	assert.Equal(t, "groups.db", v.requiredFiles[common.CheckTypeGroup][common.StorageTypeSQLite])
	assert.Equal(t, "groups.bin", v.requiredFiles[common.CheckTypeGroup][common.StorageTypeBinary])
	assert.Equal(t, "groups.csv", v.requiredFiles[common.CheckTypeGroup][common.StorageTypeCSV])
}

func TestValidator_GetExportDirs(t *testing.T) {
	// Create temporary test directory structure
	tempDir := t.TempDir()

	// Create test directories and files
	testDirs := []string{
		filepath.Join(tempDir, "dir1"),
		filepath.Join(tempDir, "dir2"),
		filepath.Join(tempDir, "dir3"),
		filepath.Join(tempDir, "empty"),
	}

	for _, dir := range testDirs {
		err := os.MkdirAll(dir, 0o755)
		require.NoError(t, err)
	}

	// Create test files
	testFiles := map[string]bool{
		filepath.Join(testDirs[0], "users.db"):   true,  // dir1: SQLite export
		filepath.Join(testDirs[1], "users.bin"):  true,  // dir2: Binary export
		filepath.Join(testDirs[2], "groups.csv"): true,  // dir3: CSV export
		filepath.Join(testDirs[3], "other.txt"):  false, // empty: No export files
	}

	for file, isExport := range testFiles {
		f, err := os.Create(file)
		require.NoError(t, err)
		f.Close()
		if isExport {
			require.FileExists(t, file)
		}
	}

	// Change to temp directory for test
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
	}()

	v := NewValidator()
	dirs, err := v.GetExportDirs(".")
	require.NoError(t, err)

	// Verify that only directories with export files are returned
	assert.Len(t, dirs, 3)
	assert.Contains(t, dirs, "dir1")
	assert.Contains(t, dirs, "dir2")
	assert.Contains(t, dirs, "dir3")
	assert.NotContains(t, dirs, "empty")
}

func TestValidator_ValidateExportDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	files := []string{
		"users.db",
		"users.bin",
		"users.csv",
		"groups.db",
		"groups.bin",
		"groups.csv",
	}

	for _, file := range files {
		f, err := os.Create(filepath.Join(tempDir, file))
		require.NoError(t, err)
		f.Close()
	}

	tests := []struct {
		name        string
		checkType   common.CheckType
		storageType common.StorageType
		wantError   bool
	}{
		{
			name:        "Valid user SQLite",
			checkType:   common.CheckTypeUser,
			storageType: common.StorageTypeSQLite,
			wantError:   false,
		},
		{
			name:        "Valid user binary",
			checkType:   common.CheckTypeUser,
			storageType: common.StorageTypeBinary,
			wantError:   false,
		},
		{
			name:        "Valid user CSV",
			checkType:   common.CheckTypeUser,
			storageType: common.StorageTypeCSV,
			wantError:   false,
		},
		{
			name:        "Valid group SQLite",
			checkType:   common.CheckTypeGroup,
			storageType: common.StorageTypeSQLite,
			wantError:   false,
		},
		{
			name:        "Valid group binary",
			checkType:   common.CheckTypeGroup,
			storageType: common.StorageTypeBinary,
			wantError:   false,
		},
		{
			name:        "Valid group CSV",
			checkType:   common.CheckTypeGroup,
			storageType: common.StorageTypeCSV,
			wantError:   false,
		},
		{
			name:        "Missing file",
			checkType:   common.CheckTypeUser,
			storageType: common.StorageTypeSQLite,
			wantError:   true,
		},
	}

	v := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := tempDir
			if tt.wantError {
				testDir = filepath.Join(tempDir, "nonexistent")
			}

			err := v.ValidateExportDir(testDir, tt.checkType, tt.storageType)
			if tt.wantError {
				assert.ErrorIs(t, err, ErrMissingFile)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
