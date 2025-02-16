package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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
func (c *Checker) Check(checkType common.CheckType, id string) (*common.CheckResult, error) {
	// Determine filename based on check type
	filename := "users.csv"
	if checkType == common.CheckTypeGroup {
		filename = "groups.csv"
	}

	// Open file
	file, err := os.Open(filepath.Join(c.dir, filename))
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read header", ErrInvalidFormat)
	}

	// Validate header
	if err := validateHeader(header); err != nil {
		return nil, err
	}

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	// Check each record
	for _, record := range records {
		if len(record) != 4 {
			return nil, fmt.Errorf("%w: incorrect number of columns", ErrInvalidFormat)
		}
		if record[0] == id {
			confidence, err := strconv.ParseFloat(record[3], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid confidence value: %w", err)
			}

			result := common.CheckResult{
				Found:      true,
				Status:     record[1],
				Reason:     record[2],
				Confidence: confidence,
			}
			return &result, nil
		}
	}

	return &common.CheckResult{}, nil
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
		if len(record) != 4 {
			return 0, fmt.Errorf("%w: incorrect number of columns", ErrInvalidFormat)
		}
	}

	return uint64(len(records)), nil
}

// validateHeader checks if the CSV file has the correct header format.
func validateHeader(header []string) error {
	if len(header) != 4 || header[0] != "hash" || header[1] != "status" ||
		header[2] != "reason" || header[3] != "confidence" {
		return fmt.Errorf("%w: expected header 'hash,status,reason,confidence'", ErrInvalidFormat)
	}
	return nil
}
