package csv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/robalyx/rotten/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestFiles(t *testing.T, dir string) {
	files := map[string][]string{
		"users.csv": {
			"hash,status,reason",
			"testHash123,banned,violation",
		},
		"groups.csv": {
			"hash,status,reason",
		},
	}

	for filename, lines := range files {
		content := ""
		for i, line := range lines {
			content += line
			if i < len(lines)-1 {
				content += "\n"
			}
		}
		err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o600)
		require.NoError(t, err)
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

func TestChecker_InvalidFileFormat(t *testing.T) {
	tempDir := t.TempDir()

	// Create file with invalid format (no header)
	err := os.WriteFile(filepath.Join(tempDir, "users.csv"), []byte("invalid,data,row\n"), 0o600)
	require.NoError(t, err)

	checker := New(tempDir)

	// Test Check with invalid file format
	found, status, reason, err := checker.Check(common.CheckTypeUser, "0123456789abcdef")
	assert.Error(t, err)
	assert.False(t, found)
	assert.Empty(t, status)
	assert.Empty(t, reason)

	// Test GetHashCount with invalid file format
	count, err := checker.GetHashCount(common.CheckTypeUser)
	assert.Error(t, err)
	assert.Zero(t, count)
}

func TestChecker_MalformedCSV(t *testing.T) {
	tempDir := t.TempDir()

	// Create file with malformed CSV
	malformedContent := "hash,status,reason\n\"unclosed quote,bad,data\nmore,bad,data"
	err := os.WriteFile(filepath.Join(tempDir, "users.csv"), []byte(malformedContent), 0o600)
	require.NoError(t, err)

	checker := New(tempDir)

	// Test Check with malformed CSV
	found, status, reason, err := checker.Check(common.CheckTypeUser, "0123456789abcdef")
	assert.Error(t, err)
	assert.False(t, found)
	assert.Empty(t, status)
	assert.Empty(t, reason)

	// Test GetHashCount with malformed CSV
	count, err := checker.GetHashCount(common.CheckTypeUser)
	assert.Error(t, err)
	assert.Zero(t, count)
}

func TestChecker_IncorrectColumnCount(t *testing.T) {
	tempDir := t.TempDir()

	// Create file with wrong number of columns
	content := "hash,status,reason\ntestHash123,banned" // Missing reason column
	err := os.WriteFile(filepath.Join(tempDir, "users.csv"), []byte(content), 0o600)
	require.NoError(t, err)

	checker := New(tempDir)

	// Test Check with incorrect column count
	found, status, reason, err := checker.Check(common.CheckTypeUser, "testHash123")
	assert.Error(t, err)
	assert.False(t, found)
	assert.Empty(t, status)
	assert.Empty(t, reason)
}

func TestChecker_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create empty file
	err := os.WriteFile(filepath.Join(tempDir, "users.csv"), []byte(""), 0o600)
	require.NoError(t, err)

	checker := New(tempDir)

	// Test Check with empty file
	found, status, reason, err := checker.Check(common.CheckTypeUser, "0123456789abcdef")
	assert.Error(t, err)
	assert.False(t, found)
	assert.Empty(t, status)
	assert.Empty(t, reason)

	// Test GetHashCount with empty file
	count, err := checker.GetHashCount(common.CheckTypeUser)
	assert.Error(t, err)
	assert.Zero(t, count)
}
