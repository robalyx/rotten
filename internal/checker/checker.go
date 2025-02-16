package checker

import (
	"errors"
	"fmt"

	"github.com/robalyx/rotten/internal/checker/binary"
	"github.com/robalyx/rotten/internal/checker/csv"
	"github.com/robalyx/rotten/internal/checker/sqlite"
	"github.com/robalyx/rotten/internal/common"
)

var ErrUnsupportedStorageType = errors.New("unsupported storage type")

// Checker interface defines the methods required for checking IDs.
type Checker interface {
	Check(checkType common.CheckType, id string) (*common.CheckResult, error)
	GetHashCount(checkType common.CheckType) (uint64, error)
}

// New creates a new checker instance based on the storage type.
func New(dir string, storageType common.StorageType) (Checker, error) {
	switch storageType {
	case common.StorageTypeSQLite:
		return sqlite.New(dir), nil
	case common.StorageTypeBinary:
		return binary.New(dir), nil
	case common.StorageTypeCSV:
		return csv.New(dir), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedStorageType, storageType)
	}
}
