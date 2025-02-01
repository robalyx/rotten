package exports

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	d := New("owner", "repo")
	assert.NotNil(t, d)
	assert.NotNil(t, d.client)
	assert.Equal(t, "owner", d.owner)
	assert.Equal(t, "repo", d.repo)
}

func TestDownloader_ExtractZip(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "test.zip")

	// Create a test zip file
	createTestZip(t, zipPath)

	// Create destination directory
	destDir := filepath.Join(tempDir, "extracted")

	d := New("owner", "repo")
	err := d.extractZip(zipPath, destDir)
	require.NoError(t, err)

	// Verify extracted files
	assertFileExists(t, filepath.Join(destDir, "test.txt"))
	assertFileExists(t, filepath.Join(destDir, "subdir", "test2.txt"))
	assertFileContents(t, filepath.Join(destDir, "test.txt"), "test content")
	assertFileContents(t, filepath.Join(destDir, "subdir", "test2.txt"), "test content 2")
}

func TestDownloader_ExtractZip_ZipSlipProtection(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "malicious.zip")

	// Create a zip file with a malicious path
	w, err := os.Create(zipPath)
	require.NoError(t, err)
	defer w.Close()

	zw := zip.NewWriter(w)
	_, err = zw.Create("../../../etc/passwd")
	require.NoError(t, err)
	zw.Close()

	// Try to extract
	destDir := filepath.Join(tempDir, "extracted")
	d := New("owner", "repo")
	err = d.extractZip(zipPath, destDir)
	require.NoError(t, err) // Should not error, just skip malicious files

	// Verify no files were extracted outside destination
	assert.NoDirExists(t, filepath.Join(tempDir, "etc"))
}

func TestDownloader_VerifyChecksum(t *testing.T) {
	// Create a temporary file with known content
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.zip")
	content := []byte("test content")
	err := os.WriteFile(filePath, content, 0o600)
	require.NoError(t, err)

	// Get actual checksum of the test content
	h := sha256.New()
	h.Write(content)
	actualChecksum := hex.EncodeToString(h.Sum(nil))

	tests := []struct {
		name          string
		releaseNotes  string
		assetName     string
		expectedError error
		modifyContent bool
	}{
		{
			name: "Valid checksum",
			releaseNotes: fmt.Sprintf(`## Test Release
test.zip SHA256: %s`, actualChecksum),
			assetName:     "test.zip",
			expectedError: nil,
		},
		{
			name: "Invalid checksum",
			releaseNotes: `## Test Release
test.zip SHA256: 0000000000000000000000000000000000000000000000000000000000000000`,
			assetName:     "test.zip",
			expectedError: ErrChecksumMismatch,
		},
		{
			name: "No checksum in notes",
			releaseNotes: `## Test Release
Some other content`,
			assetName:     "test.zip",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			release := &Release{
				Body: tt.releaseNotes,
			}
			asset := &Asset{
				Name: tt.assetName,
			}

			d := New("owner", "repo")
			err := d.verifyChecksum(release, asset, filePath)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createTestZip(t *testing.T, zipPath string) {
	t.Helper()

	w, err := os.Create(zipPath)
	require.NoError(t, err)
	defer w.Close()

	zw := zip.NewWriter(w)
	defer zw.Close()

	// Add test.txt
	f1, err := zw.Create("test.txt")
	require.NoError(t, err)
	_, err = f1.Write([]byte("test content"))
	require.NoError(t, err)

	// Add subdir/test2.txt
	f2, err := zw.Create("subdir/test2.txt")
	require.NoError(t, err)
	_, err = f2.Write([]byte("test content 2"))
	require.NoError(t, err)
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.NoError(t, err, "file should exist: %s", path)
}

func assertFileContents(t *testing.T, path, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, expected, string(content))
}
