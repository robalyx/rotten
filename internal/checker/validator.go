package checker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/robalyx/rotten/internal/common"
)

const (
	// MaxSearchDepth is the maximum directory depth to search for exports.
	MaxSearchDepth = 3
)

var ErrMissingFile = errors.New("missing required file")

// Validator handles validation of export directories and files.
type Validator struct {
	requiredFiles map[common.CheckType]map[common.StorageType]string
}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return &Validator{
		requiredFiles: map[common.CheckType]map[common.StorageType]string{
			common.CheckTypeUser: {
				common.StorageTypeSQLite: "users.db",
				common.StorageTypeBinary: "users.bin",
				common.StorageTypeCSV:    "users.csv",
			},
			common.CheckTypeGroup: {
				common.StorageTypeSQLite: "groups.db",
				common.StorageTypeBinary: "groups.bin",
				common.StorageTypeCSV:    "groups.csv",
			},
		},
	}
}

// GetExportDirs returns a list of valid export directories.
func (v *Validator) GetExportDirs(baseDir string) ([]string, error) {
	dirs := make(map[string]struct{})
	validFiles := make(map[string]struct{})

	// Build valid files map
	for _, files := range v.requiredFiles {
		for _, filename := range files {
			validFiles[filename] = struct{}{}
		}
	}

	// Check if directory exists first
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Find all valid export directories
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate current depth relative to base directory
		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		depth := len(strings.Split(relPath, string(filepath.Separator)))

		// Skip if we've exceeded max depth
		if depth > MaxSearchDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			if _, ok := validFiles[filepath.Base(path)]; ok {
				dirs[filepath.Dir(path)] = struct{}{}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directories: %w", err)
	}

	// Convert map keys to slice
	result := make([]string, 0, len(dirs))
	for dir := range dirs {
		result = append(result, dir)
	}

	return result, nil
}

// ValidateExportDir ensures required files exist in the directory for the given storage type.
func (v *Validator) ValidateExportDir(dir string, checkType common.CheckType, storageType common.StorageType) error {
	filename := v.requiredFiles[checkType][storageType]
	path := filepath.Join(dir, filename)
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("%w: %s file for %s check: %s", ErrMissingFile, storageType, checkType, filename)
	}
	return nil
}
