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
func (c *Checker) Check(checkType common.CheckType, id string) (*common.CheckResult, error) {
	// Convert input ID to bytes
	searchHash, err := hex.DecodeString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid hash format: %w", err)
	}

	// Open and validate file
	file, count, err := c.openAndValidateFile(checkType)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read and compare each record
	hashBuf := make([]byte, len(searchHash))
	for range count {
		// Read and compare hash
		if found, result, err := c.readAndCompareHash(file, hashBuf, searchHash); err != nil {
			return nil, err
		} else if found {
			return result, nil
		}

		// Skip remaining record data
		if err := c.skipRecordData(file); err != nil {
			return nil, err
		}
	}

	return &common.CheckResult{}, nil
}

// openAndValidateFile opens the binary file and validates its format.
func (c *Checker) openAndValidateFile(checkType common.CheckType) (*os.File, uint32, error) {
	// Determine filename based on check type
	filename := "users.bin"
	if checkType == common.CheckTypeGroup {
		filename = "groups.bin"
	}

	// Open file
	file, err := os.Open(filepath.Join(c.dir, filename))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open file: %w", err)
	}

	// Validate file format
	count, err := c.validateFileFormat(file)
	if err != nil {
		file.Close()
		return nil, 0, err
	}

	return file, count, nil
}

// validateFileFormat checks the file size and reads the record count.
func (c *Checker) validateFileFormat(file *os.File) (uint32, error) {
	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Validate minimum file size
	minFileSize := int64(4) // minimum size for count
	if stat.Size() < minFileSize {
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

	return count, nil
}

// readAndCompareHash reads a hash from the file and compares it with the search hash.
func (c *Checker) readAndCompareHash(file *os.File, hashBuf, searchHash []byte) (bool, *common.CheckResult, error) {
	// Read hash
	if _, err := io.ReadFull(file, hashBuf); err != nil {
		return false, nil, fmt.Errorf("failed to read hash: %w", err)
	}

	// If hash matches, read status and reason
	if string(hashBuf) == string(searchHash) {
		result, err := c.readRecordData(file)
		if err != nil {
			return false, nil, err
		}
		return true, result, nil
	}

	return false, nil, nil
}

// readRecordData reads the status, reason, and confidence for a matching record.
func (c *Checker) readRecordData(file *os.File) (*common.CheckResult, error) {
	var result common.CheckResult
	result.Found = true

	// Read status
	statusBuf, err := c.readLengthAndData(file, "status")
	if err != nil {
		return nil, err
	}
	result.Status = string(statusBuf)

	// Read reason
	reasonBuf, err := c.readLengthAndData(file, "reason")
	if err != nil {
		return nil, err
	}
	result.Reason = string(reasonBuf)

	// Read confidence
	if err := binary.Read(file, binary.LittleEndian, &result.Confidence); err != nil {
		return nil, fmt.Errorf("failed to read confidence: %w", err)
	}

	return &result, nil
}

// readLengthAndData reads a length-prefixed string from the file.
func (c *Checker) readLengthAndData(file *os.File, fieldName string) ([]byte, error) {
	var length uint16
	if err := binary.Read(file, binary.LittleEndian, &length); err != nil {
		return nil, fmt.Errorf("failed to read %s length: %w", fieldName, err)
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(file, buf); err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", fieldName, err)
	}

	return buf, nil
}

// skipRecordData skips over the status, reason, and confidence fields.
func (c *Checker) skipRecordData(file *os.File) error {
	var skipLen uint16
	// Skip status
	if err := binary.Read(file, binary.LittleEndian, &skipLen); err != nil {
		return fmt.Errorf("failed to read status length: %w", err)
	}
	if _, err := file.Seek(int64(skipLen), io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to skip status: %w", err)
	}

	// Skip reason and confidence
	if err := binary.Read(file, binary.LittleEndian, &skipLen); err != nil {
		return fmt.Errorf("failed to read reason length: %w", err)
	}
	if _, err := file.Seek(int64(skipLen)+8, io.SeekCurrent); err != nil { // +8 for confidence float64
		return fmt.Errorf("failed to skip reason and confidence: %w", err)
	}

	return nil
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
