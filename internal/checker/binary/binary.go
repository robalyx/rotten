package binary

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/robalyx/rotten/internal/common"
)

var ErrInvalidFormat = errors.New("invalid file format")

// Checker implements the common.Checker interface for binary storage.
type Checker struct {
	dir string
}

// New creates a new binary checker.
func New(dir string) *Checker {
	return &Checker{dir: dir}
}

// Check verifies if the given ID exists in the binary file.
func (c *Checker) Check(checkType common.CheckType, id string) (bool, string, string, error) {
	// Determine filename based on check type
	filename := "users.bin"
	if checkType == common.CheckTypeGroup {
		filename = "groups.bin"
	}

	// Convert input ID to bytes
	searchHash, err := hex.DecodeString(id)
	if err != nil {
		return false, "", "", fmt.Errorf("invalid hash format: %w", err)
	}

	// Open file
	file, err := os.Open(filepath.Join(c.dir, filename))
	if err != nil {
		return false, "", "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return false, "", "", fmt.Errorf("failed to get file stats: %w", err)
	}

	// Validate minimum file size
	minFileSize := int64(4) // minimum size for count
	if stat.Size() < minFileSize {
		return false, "", "", fmt.Errorf("%w: file too small", ErrInvalidFormat)
	}

	// Read count of hashes
	var count uint32
	if err := binary.Read(file, binary.LittleEndian, &count); err != nil {
		return false, "", "", fmt.Errorf("%w: failed to read count", ErrInvalidFormat)
	}

	// Validate count against file size
	minRecordSize := len("0123456789abcdef") + 4 // hash length + 4 bytes for lengths
	expectedMinSize := 4 + (int64(count) * int64(minRecordSize))
	if stat.Size() < expectedMinSize {
		return false, "", "", fmt.Errorf("%w: file size too small for count", ErrInvalidFormat)
	}

	// Read and compare each record
	hashBuf := make([]byte, len(searchHash))
	for range count {
		// Read hash
		_, err := io.ReadFull(file, hashBuf)
		if err != nil {
			return false, "", "", fmt.Errorf("failed to read hash: %w", err)
		}

		// If hash matches, read status and reason
		if string(hashBuf) == string(searchHash) {
			// Read status length and string
			var statusLen uint16
			if err := binary.Read(file, binary.LittleEndian, &statusLen); err != nil {
				return false, "", "", fmt.Errorf("failed to read status length: %w", err)
			}
			statusBuf := make([]byte, statusLen)
			if _, err := io.ReadFull(file, statusBuf); err != nil {
				return false, "", "", fmt.Errorf("failed to read status: %w", err)
			}

			// Read reason length and string
			var reasonLen uint16
			if err := binary.Read(file, binary.LittleEndian, &reasonLen); err != nil {
				return false, "", "", fmt.Errorf("failed to read reason length: %w", err)
			}
			reasonBuf := make([]byte, reasonLen)
			if _, err := io.ReadFull(file, reasonBuf); err != nil {
				return false, "", "", fmt.Errorf("failed to read reason: %w", err)
			}

			return true, string(statusBuf), string(reasonBuf), nil
		}

		// Skip status and reason for non-matching hash
		var skipLen uint16
		// Skip status
		if err := binary.Read(file, binary.LittleEndian, &skipLen); err != nil {
			return false, "", "", fmt.Errorf("failed to read status length: %w", err)
		}
		if _, err := file.Seek(int64(skipLen), io.SeekCurrent); err != nil {
			return false, "", "", fmt.Errorf("failed to skip status: %w", err)
		}
		// Skip reason
		if err := binary.Read(file, binary.LittleEndian, &skipLen); err != nil {
			return false, "", "", fmt.Errorf("failed to read reason length: %w", err)
		}
		if _, err := file.Seek(int64(skipLen), io.SeekCurrent); err != nil {
			return false, "", "", fmt.Errorf("failed to skip reason: %w", err)
		}
	}

	return false, "", "", nil
}

// GetHashCount returns the number of hashes in the binary file.
func (c *Checker) GetHashCount(checkType common.CheckType) (uint64, error) {
	// Determine filename based on check type
	filename := "users.bin"
	if checkType == common.CheckTypeGroup {
		filename = "groups.bin"
	}

	// Open file
	file, err := os.Open(filepath.Join(c.dir, filename))
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Check minimum file size
	if stat.Size() < 4 {
		return 0, fmt.Errorf("%w: file too small", ErrInvalidFormat)
	}

	// Read count of hashes
	var count uint32
	if err := binary.Read(file, binary.LittleEndian, &count); err != nil {
		return 0, fmt.Errorf("%w: failed to read count", ErrInvalidFormat)
	}

	// Validate count against file size
	minRecordSize := len("0123456789abcdef") + 4 // hash length + 4 bytes for lengths
	expectedMinSize := 4 + (int64(count) * int64(minRecordSize))
	if stat.Size() < expectedMinSize {
		return 0, fmt.Errorf("%w: file size too small for count", ErrInvalidFormat)
	}

	return uint64(count), nil
}
