package tui

// State represents the current state of the TUI.
type State int

const (
	// StateCheckType is the initial state where user selects between user/group/friends check.
	StateCheckType State = iota
	// StateDirectory is the state where user selects an export directory.
	StateDirectory
	// StateStorageType is the state where user selects storage type (SQLite/Binary/CSV).
	StateStorageType
	// StateIDInput is the state where user enters an ID to check.
	StateIDInput
	// StateUserGroupResult is the state where user/group check results are displayed.
	StateUserGroupResult
	// StateFriendsResult is the state where friend check results are displayed.
	StateFriendsResult
	// StateExportDownload is the state where user can download official exports.
	StateExportDownload
)
