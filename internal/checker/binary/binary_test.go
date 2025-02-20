package binary

import (
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/robalyx/rotten/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestFiles(t *testing.T, dir string) {
	files := []string{
		filepath.Join(dir, "users.bin"),
		filepath.Join(dir, "groups.bin"),
	}

	for _, file := range files {
		f, err := os.Create(file)
		require.NoError(t, err)
		defer f.Close()

		// Write count (0) in little-endian
		err = binary.Write(f, binary.LittleEndian, uint32(0))
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := New(tempDir)
			result, err := checker.Check(tt.checkType, tt.hash)

			if tt.wantErrType != nil {
				assert.ErrorIs(t, err, tt.wantErrType)
				return
			}

			if tt.wantFound {
				assert.NoError(t, err)
				assert.True(t, result.Found)
				assert.Equal(t, tt.wantStatus, result.Status)
				assert.Equal(t, tt.wantReason, result.Reason)
			} else {
				assert.False(t, result.Found)
				assert.Empty(t, result.Status)
				assert.Empty(t, result.Reason)
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
			wantCount: 0,
		},
		{
			name:      "Valid group check",
			checkType: common.CheckTypeGroup,
			wantCount: 0,
		},
		{
			name:      "Valid friends check",
			checkType: common.CheckTypeFriends,
			wantCount: 0,
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
	result, err := checker.Check(common.CheckTypeUser, "0123456789abcdef")
	assert.Error(t, err)
	assert.Nil(t, result)

	// Test GetHashCount with nonexistent file
	count, err := checker.GetHashCount(common.CheckTypeUser)
	assert.Error(t, err)
	assert.Zero(t, count)
}

func TestChecker_InvalidFileFormat(t *testing.T) {
	tempDir := t.TempDir()

	// Create file with invalid format
	err := os.WriteFile(filepath.Join(tempDir, "users.bin"), []byte("invalid"), 0o600)
	require.NoError(t, err)

	checker := New(tempDir)

	// Test Check with invalid file format
	result, err := checker.Check(common.CheckTypeUser, "0123456789abcdef")
	assert.Error(t, err)
	assert.Nil(t, result)

	// Test GetHashCount with invalid file format
	count, err := checker.GetHashCount(common.CheckTypeUser)
	assert.Error(t, err)
	assert.Zero(t, count)
}

func TestChecker_WithData(t *testing.T) {
	tempDir := t.TempDir()
	testHash := "0123456789abcdef"
	testStatus := "banned"
	testReason := "violation"
	testConfidence := float64(0.95)

	// Create test file with one record
	f, err := os.Create(filepath.Join(tempDir, "users.bin"))
	require.NoError(t, err)
	defer f.Close()

	// Write count (1)
	err = binary.Write(f, binary.LittleEndian, uint32(1))
	require.NoError(t, err)

	// Write hash
	hashBytes, err := hex.DecodeString(testHash)
	require.NoError(t, err)
	_, err = f.Write(hashBytes)
	require.NoError(t, err)

	// Write status
	err = binary.Write(f, binary.LittleEndian, uint16(len(testStatus)))
	require.NoError(t, err)
	_, err = f.Write([]byte(testStatus))
	require.NoError(t, err)

	// Write reason
	err = binary.Write(f, binary.LittleEndian, uint16(len(testReason)))
	require.NoError(t, err)
	_, err = f.Write([]byte(testReason))
	require.NoError(t, err)

	// Write confidence
	err = binary.Write(f, binary.LittleEndian, testConfidence)
	require.NoError(t, err)

	checker := New(tempDir)

	// Test finding the record
	result, err := checker.Check(common.CheckTypeUser, testHash)
	assert.NoError(t, err)
	assert.True(t, result.Found)
	assert.Equal(t, testStatus, result.Status)
	assert.Equal(t, testReason, result.Reason)
	assert.Equal(t, testConfidence, result.Confidence)

	// Test count
	count, err := checker.GetHashCount(common.CheckTypeUser)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), count)
}

func TestChecker_InvalidHashFormat(t *testing.T) {
	tempDir := t.TempDir()
	checker := New(tempDir)

	result, err := checker.Check(common.CheckTypeUser, "invalid")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid hash format")
}
