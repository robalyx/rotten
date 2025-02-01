//nolint:tagliatelle
package exports

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/robalyx/rotten/internal/version"
)

var (
	ErrNewerVersionAvailable = errors.New("newer version of Rotten is available - please update your program")
	ErrChecksumMismatch      = errors.New("checksum mismatch")
	ErrZipSlipAttack         = errors.New("zip slip attack detected")
	ErrInvalidStatusCode     = errors.New("invalid status code")
	ErrDownloadFailed        = errors.New("download failed")
)

// maxDecompressedSize is the maximum size of decompressed data (1GB).
const maxDecompressedSize = 1 << 30

// ProgressWriter wraps an io.Writer to track progress.
type ProgressWriter struct {
	w io.Writer
}

// Write implements io.Writer.
func (pw *ProgressWriter) Write(p []byte) (int, error) {
	return pw.w.Write(p)
}

// Release represents a GitHub release.
type Release struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	Assets      []Asset `json:"assets"`
	DownloadURL string  `json:"html_url"`
}

// Asset represents a downloadable file in a release.
type Asset struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

// Downloader handles downloading and verifying export files.
type Downloader struct {
	client *http.Client
	owner  string
	repo   string
}

// New creates a new Downloader instance.
func New(owner, repo string) *Downloader {
	return &Downloader{
		client: http.DefaultClient,
		owner:  owner,
		repo:   repo,
	}
}

// GetAvailableExports returns a list of available export releases.
func (d *Downloader) GetAvailableExports(ctx context.Context) ([]*Release, error) {
	// Get releases from GitHub API
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", d.owner, d.repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrInvalidStatusCode, resp.StatusCode)
	}

	var releases []*Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}

	// Parse current version from build-time variable
	currentVersion, err := version.Parse(version.EngineVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current version: %w", err)
	}

	// Filter compatible releases and check for newer versions
	compatibleReleases := make([]*Release, 0)
	hasNewerVersion := false

	for _, release := range releases {
		// Extract engine version from release notes
		engineVersion := version.ExtractFromNotes(release.Body)
		if engineVersion == "" {
			continue // Skip releases without engine version
		}

		releaseVersion, err := version.Parse(engineVersion)
		if err != nil {
			continue // Skip releases with invalid version
		}

		if currentVersion.IsCompatible(releaseVersion) {
			compatibleReleases = append(compatibleReleases, release)
		} else if currentVersion.IsNewer(releaseVersion) {
			hasNewerVersion = true
		}
	}

	if hasNewerVersion {
		return compatibleReleases, ErrNewerVersionAvailable
	}

	return compatibleReleases, nil
}

// DownloadExport downloads and extracts an export release.
func (d *Downloader) DownloadExport(ctx context.Context, release *Release, destDir string) error {
	// Clean up any existing export directory
	if err := os.RemoveAll(destDir); err != nil {
		return fmt.Errorf("failed to clean existing export directory: %w", err)
	}

	// Create temporary directory for download
	tmpDir, err := os.MkdirTemp("", "rotten-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download each asset
	for _, asset := range release.Assets {
		if !strings.HasSuffix(asset.Name, ".zip") {
			continue
		}

		// Download zip file
		zipPath := filepath.Join(tmpDir, asset.Name)
		if err := d.downloadAsset(ctx, &asset, zipPath); err != nil {
			return fmt.Errorf("failed to download asset: %w", err)
		}

		// Verify checksum if available
		if err := d.verifyChecksum(release, &asset, zipPath); err != nil {
			return fmt.Errorf("failed to verify checksum: %w", err)
		}

		// Extract zip file to temp directory first
		extractDir := filepath.Join(tmpDir, "extracted")
		if err := d.extractZip(zipPath, extractDir); err != nil {
			return fmt.Errorf("failed to extract zip: %w", err)
		}

		// Remove any existing destination directory
		if err := os.RemoveAll(destDir); err != nil {
			return fmt.Errorf("failed to clean destination directory: %w", err)
		}

		// Create destination directory
		if err := os.MkdirAll(filepath.Dir(destDir), 0o755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}

		// Move extracted contents to final destination
		if err := os.Rename(extractDir, destDir); err != nil {
			return fmt.Errorf("failed to move extracted files to destination: %w", err)
		}
	}

	return nil
}

// downloadAsset downloads a single release asset to the specified path.
func (d *Downloader) downloadAsset(ctx context.Context, asset *Asset, destPath string) error {
	// Download the asset
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.DownloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d", ErrDownloadFailed, resp.StatusCode)
	}

	// Create destination file
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	pw := &ProgressWriter{w: f}
	if _, err := io.Copy(pw, resp.Body); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// verifyChecksum verifies the SHA256 checksum of a downloaded file against the release notes.
func (d *Downloader) verifyChecksum(release *Release, asset *Asset, filePath string) error {
	// Look for checksum in release body
	checksumPrefix := asset.Name + " SHA256: "
	if !strings.Contains(release.Body, checksumPrefix) {
		return nil // No checksum available
	}

	// Extract expected checksum
	// Find the line containing the checksum
	lines := strings.Split(release.Body, "\n")
	var expectedChecksum string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, checksumPrefix) {
			expectedChecksum = strings.TrimPrefix(line, checksumPrefix)
			expectedChecksum = strings.TrimSpace(expectedChecksum)
			break
		}
	}

	if len(expectedChecksum) != 64 {
		return fmt.Errorf("%w: invalid checksum format", ErrChecksumMismatch)
	}

	// Calculate actual checksum
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := hex.EncodeToString(h.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("%w: %s", ErrChecksumMismatch, asset.Name)
	}

	return nil
}

// extractZip extracts a zip file to the specified destination directory.
func (d *Downloader) extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		// Clean the file path and ensure it's safe
		fileName := filepath.Clean(f.Name)
		if strings.HasPrefix(fileName, "../") || strings.HasPrefix(fileName, "/") {
			continue // Skip files that try to escape the destination directory
		}

		// Create directory for file
		fpath := filepath.Join(destDir, fileName)
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directories: %w", err)
		}

		// Skip if directory
		if f.FileInfo().IsDir() {
			continue
		}

		// Create file
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		// Extract file with size limit
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("failed to open zip file: %w", err)
		}

		limitedReader := io.LimitReader(rc, maxDecompressedSize)
		if _, err := io.Copy(outFile, limitedReader); err != nil {
			outFile.Close()
			rc.Close()
			return fmt.Errorf("failed to extract file: %w", err)
		}

		outFile.Close()
		rc.Close()
	}

	return nil
}
