package common

// CheckType represents the type of check to perform.
type CheckType string

const (
	CheckTypeUser    CheckType = "user"
	CheckTypeGroup   CheckType = "group"
	CheckTypeFriends CheckType = "friends"
)

// StorageType represents the type of storage medium.
type StorageType string

const (
	StorageTypeSQLite StorageType = "sqlite"
	StorageTypeBinary StorageType = "binary"
	StorageTypeCSV    StorageType = "csv"
)
