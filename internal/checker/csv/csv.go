package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robalyx/rotten/internal/common"
)

var ErrInvalidFormat = errors.New("invalid CSV format")

// Checker implements the common.Checker interface for CSV storage.
type Checker struct {
	dir string
}

// Result contains the check result details.
type Result struct {
	Found  bool
	Status string
	Reason string
}

// New creates a new CSV checker.
func New(dir string) *Checker {
	return &Checker{dir: dir}
}

// Check verifies if the given ID exists in the CSV file.
func (c *Checker) Check(checkType common.CheckType, id string) (bool, string, string, error) {
	// Determine filename based on check type
	filename := "users.csv"
	if checkType == common.CheckTypeGroup {
		filename = "groups.csv"
	}

	// Open file
	file, err := os.Open(filepath.Join(c.dir, filename))
	if err != nil {
		return false, "", "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return false, "", "", fmt.Errorf("%w: failed to read header", ErrInvalidFormat)
	}

	// Validate header
	if err := validateHeader(header); err != nil {
		return false, "", "", err
	}

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return false, "", "", fmt.Errorf("failed to read CSV: %w", err)
	}

	// Check each record
	for _, record := range records {
		if len(record) != 3 {
			return false, "", "", fmt.Errorf("%w: incorrect number of columns", ErrInvalidFormat)
		}
		if record[0] == id {
			return true, record[1], record[2], nil
		}
	}

	return false, "", "", nil
}

// GetHashCount returns the number of hashes in the CSV file.
func (c *Checker) GetHashCount(checkType common.CheckType) (uint64, error) {
	// Determine filename based on check type
	filename := "users.csv"
	if checkType == common.CheckTypeGroup {
		filename = "groups.csv"
	}

	// Open file
	file, err := os.Open(filepath.Join(c.dir, filename))
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("%w: failed to read header", ErrInvalidFormat)
	}

	// Validate header
	if err := validateHeader(header); err != nil {
		return 0, err
	}

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("failed to read CSV: %w", err)
	}

	// Validate record format
	for _, record := range records {
		if len(record) != 3 {
			return 0, fmt.Errorf("%w: incorrect number of columns", ErrInvalidFormat)
		}
	}

	return uint64(len(records)), nil
}

// validateHeader checks if the CSV file has the correct header format.
func validateHeader(header []string) error {
	if len(header) != 3 || header[0] != "hash" || header[1] != "status" || header[2] != "reason" {
		return fmt.Errorf("%w: expected header 'hash,status,reason'", ErrInvalidFormat)
	}
	return nil
}
