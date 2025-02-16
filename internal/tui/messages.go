package tui

import "github.com/robalyx/rotten/internal/exports"

// ExportsLoadedMsg is sent when exports are loaded from GitHub.
type ExportsLoadedMsg struct {
	Releases []*exports.Release
	Error    error
}

// ExportDownloadCompleteMsg is sent when an export download finishes.
type ExportDownloadCompleteMsg struct {
	Error   error
	TempDir string
}

// CheckProgressMsg is sent to indicate progress in checking an ID.
type CheckProgressMsg struct {
	Complete   bool
	Error      error
	Found      bool
	Status     string
	Reason     string
	Confidence float64
}

// FriendsCheckProgressMsg is sent to indicate progress in checking friends.
type FriendsCheckProgressMsg struct {
	Complete      bool
	Error         error
	TotalChecked  int
	TotalFriends  int
	FlaggedCount  int
	FriendResults []FriendResult
}
