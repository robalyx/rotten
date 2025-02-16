package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaxron/roapi.go/pkg/api/resources/friends"
	"github.com/robalyx/rotten/internal/checker"
	"github.com/robalyx/rotten/internal/common"
	"github.com/robalyx/rotten/internal/config"
	"github.com/robalyx/rotten/internal/exports"
)

// Update handles UI events and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case ExportsLoadedMsg:
		// Handle completion of exports loading
		if msg.Error != nil {
			m.downloadError = msg.Error
			return m, nil
		}
		m.availableExports = msg.Releases
		return m, nil

	case ExportDownloadCompleteMsg:
		// Handle completion of export download
		m.downloading = false
		if msg.Error != nil {
			m.downloadError = msg.Error
			return m, nil
		}

		// Scan the temp directory for exports
		validator := checker.NewValidator()
		dirs, err := validator.GetExportDirs(msg.TempDir)
		if err != nil {
			os.RemoveAll(msg.TempDir)
			m.err = err
			return m, nil
		}

		// Update directories list with friendly name
		for i, dir := range dirs {
			if dir == msg.TempDir {
				dirs[i] = OfficialExportDir
				break
			}
		}
		m.directories = dirs

		// Automatically select the official export and move to storage type selection
		m.selected = 0
		m.state = StateStorageType
		return m, nil

	case CheckProgressMsg:
		// Handle completion of ID check
		m.checking = false
		if msg.Error != nil {
			m.err = msg.Error
			return m, nil
		}

		m.result = msg.Found
		m.status = msg.Status
		m.reason = msg.Reason
		m.confidence = msg.Confidence
		m.state = StateUserGroupResult
		return m, nil

	case FriendsCheckProgressMsg:
		// Handle completion of friends check
		m.checking = false
		if msg.Error != nil {
			m.err = msg.Error
			return m, nil
		}
		m.totalFriendCount = msg.TotalFriends
		m.flaggedFriendCount = msg.FlaggedCount
		m.friendResults = msg.FriendResults
		m.state = StateFriendsResult
		return m, nil
	}
	return m, nil
}

// handleKeyPress processes keyboard input and updates model state accordingly.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "r":
		// Reset state
		m = *NewModel()
		return m, nil
	case "up", "k":
		// Handle upward navigation
		return m.handleUpKey(), nil
	case "down", "j":
		// Handle downward navigation
		return m.handleDownKey(), nil
	case "enter":
		// Reset if there's an error
		if m.err != nil {
			m = *NewModel()
			return m, nil
		}
		return m.handleEnterKey()
	default:
		if m.state == StateIDInput {
			return m.handleIDInput(msg), nil
		}
	}
	return m, nil
}

// handleUpKey handles up arrow key press in menus.
func (m Model) handleUpKey() tea.Model {
	switch m.state {
	case StateCheckType:
		if m.checkTypeSelected > 0 {
			m.checkTypeSelected--
		}
	case StateDirectory:
		if m.selected > 0 {
			m.selected--
		}
	case StateStorageType:
		if m.storageTypeSelected > 0 {
			m.storageTypeSelected--
		}
	case StateExportDownload:
		if !m.downloading && m.selectedExport > 0 {
			m.selectedExport--
		}
	case StateFriendsResult:
		if m.friendsScrollPos > 0 {
			m.friendsScrollPos--
		}
	}
	return m
}

// handleDownKey handles down arrow key press in menus.
func (m Model) handleDownKey() tea.Model {
	switch m.state {
	case StateCheckType:
		if m.checkTypeSelected < 2 {
			m.checkTypeSelected++
		}
	case StateDirectory:
		if m.selected < len(m.directories) {
			m.selected++
		}
	case StateStorageType:
		if m.storageTypeSelected < 2 {
			m.storageTypeSelected++
		}
	case StateExportDownload:
		if !m.downloading && m.selectedExport < len(m.availableExports)-1 {
			m.selectedExport++
		}
	case StateFriendsResult:
		if m.friendsScrollPos < len(m.friendResults)-1 {
			m.friendsScrollPos++
		}
	}
	return m
}

// handleEnterKey processes enter key press based on current state.
func (m Model) handleEnterKey() (tea.Model, tea.Cmd) {
	switch m.state {
	case StateExportDownload:
		if m.downloading {
			return m, nil // Prevent starting new download while one is in progress
		}
		if len(m.availableExports) > 0 {
			m.downloadError = nil
			m.downloading = true
			return m, m.downloadExportCmd(m.availableExports[m.selectedExport])
		}
		return m, nil

	case StateCheckType:
		// Set check type based on selection
		switch m.checkTypeSelected {
		case 0:
			m.checkType = common.CheckTypeUser
		case 1:
			m.checkType = common.CheckTypeGroup
		case 2:
			m.checkType = common.CheckTypeFriends
		}
		m.state = StateDirectory

	case StateDirectory:
		// Handle directory selection or official export download
		if m.selected == 0 {
			m.state = StateExportDownload
			return m, m.loadExportsCmd()
		}
		m.selected--
		m.state = StateStorageType

	case StateStorageType:
		// Set storage type based on selection
		switch m.storageTypeSelected {
		case 0:
			m.storageType = common.StorageTypeSQLite
		case 1:
			m.storageType = common.StorageTypeBinary
		case 2:
			m.storageType = common.StorageTypeCSV
		}
		return m.handleStorageSelection()

	case StateIDInput:
		// Process ID input if not empty
		if len(m.id) > 0 {
			if m.checkType == common.CheckTypeFriends {
				return m.handleFriendsCheck()
			}
			return m.handleIDSubmission()
		}

	case StateUserGroupResult, StateFriendsResult:
		// Reset for new ID input
		m.state = StateIDInput
		m.id = ""
		m.result = false
		m.friendsScrollPos = 0 // Reset scroll position when exiting

	default:
		return m, nil
	}
	return m, nil
}

// handleIDInput processes keyboard input during ID entry.
func (m Model) handleIDInput(msg tea.KeyMsg) tea.Model {
	switch msg.Type {
	case tea.KeyBackspace, tea.KeyDelete:
		// Remove last character
		if len(m.id) > 0 {
			m.id = m.id[:len(m.id)-1]
		}
	case tea.KeyRunes:
		// Only allow numeric input
		for _, r := range msg.Runes {
			if r >= '0' && r <= '9' {
				m.id += string(r)
			}
		}
	}
	return m
}

// handleStorageSelection initializes the checker after storage type selection.
func (m Model) handleStorageSelection() (tea.Model, tea.Cmd) {
	// Get actual directory path
	dir := m.directories[m.selected]
	tempDir := ""

	// If using downloaded export, get temp directory path
	if dir == OfficialExportDir {
		tempDir = filepath.Join(os.TempDir(), "rotector-exports")
		dir = tempDir

		// Clean up temp directory when done
		defer func() {
			if m.err != nil {
				os.RemoveAll(tempDir)
			}
		}()
	}

	// Validate export directory
	if err := m.validator.ValidateExportDir(dir, m.checkType, m.storageType); err != nil {
		m.err = fmt.Errorf("invalid export directory: %w", err)
		return m, nil
	}

	// Load configuration
	cfg, err := config.LoadOrCreate(dir)
	if err != nil {
		m.err = fmt.Errorf("failed to load configuration: %w", err)
		return m, nil
	}
	m.config = cfg

	// Initialize checker
	m.checker, err = checker.New(dir, m.storageType)
	if err != nil {
		m.err = err
		return m, nil
	}

	// Get hash count
	m.hashCount, err = m.checker.GetHashCount(m.checkType)
	if err != nil {
		m.err = err
		return m, nil
	}

	m.state = StateIDInput
	return m, nil
}

// handleIDSubmission processes the entered ID and performs the check.
func (m Model) handleIDSubmission() (tea.Model, tea.Cmd) {
	if m.checking {
		return m, nil
	}

	// Parse and validate ID
	id, err := strconv.ParseUint(m.id, 10, 64)
	if err != nil {
		m.err = fmt.Errorf("invalid ID format: %w", err)
		return m, nil
	}

	m.checking = true
	return m, func() tea.Msg {
		// Hash ID and check against export
		hashType := HashType(m.config.HashType)
		hash := hashID(id, m.config.Salt, hashType, m.config.Iterations, m.config.Memory)

		result, err := m.checker.Check(m.checkType, hash)
		return CheckProgressMsg{
			Complete:   true,
			Error:      err,
			Found:      result != nil && result.Found,
			Status:     result.Status,
			Reason:     result.Reason,
			Confidence: result.Confidence,
		}
	}
}

// handleFriendsCheck handles the friends check.
func (m Model) handleFriendsCheck() (tea.Model, tea.Cmd) {
	if m.checking {
		return m, nil
	}

	// Parse and validate user ID
	userID, err := strconv.ParseUint(m.id, 10, 64)
	if err != nil {
		m.err = fmt.Errorf("invalid ID format: %w", err)
		return m, nil
	}

	m.checking = true
	return m, func() tea.Msg {
		var (
			cursor        string
			hasNextPage   = true
			totalChecked  = 0
			flaggedCount  = 0
			friendResults = make([]FriendResult, 0)
		)

		hashType := HashType(m.config.HashType)
		for hasNextPage {
			// Fetch page of friends
			params := friends.NewFindFriendsBuilder(userID).
				WithLimit(50).
				WithCursor(cursor).
				Build()

			friendsList, err := m.roAPI.Friends().FindFriends(context.Background(), params)
			if err != nil {
				return FriendsCheckProgressMsg{
					Complete: true,
					Error:    fmt.Errorf("failed to fetch friends: %w", err),
				}
			}

			// Update total count and process current page
			totalChecked += len(friendsList.PageItems)

			// Check each friend in current page
			for _, friend := range friendsList.PageItems {
				hash := hashID(friend.ID, m.config.Salt, hashType, m.config.Iterations, m.config.Memory)

				result, err := m.checker.Check(common.CheckTypeUser, hash)
				if err != nil {
					return FriendsCheckProgressMsg{
						Complete: true,
						Error:    fmt.Errorf("failed to check friend %d: %w", friend.ID, err),
					}
				}

				if result.Found {
					flaggedCount++
					friendResults = append(friendResults, FriendResult{
						ID:         friend.ID,
						Found:      true,
						Status:     result.Status,
						Reason:     result.Reason,
						Confidence: result.Confidence,
					})
				}
			}

			// Check if there are more pages
			if friendsList.NextCursor != nil {
				cursor = *friendsList.NextCursor
			} else {
				hasNextPage = false
			}
		}

		return FriendsCheckProgressMsg{
			Complete:      true,
			TotalChecked:  totalChecked,
			TotalFriends:  totalChecked,
			FlaggedCount:  flaggedCount,
			FriendResults: friendResults,
		}
	}
}

// loadExportsCmd creates a command to load available exports.
func (m Model) loadExportsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		releases, err := m.downloader.GetAvailableExports(ctx)
		return ExportsLoadedMsg{
			Releases: releases,
			Error:    err,
		}
	}
}

// downloadExportCmd creates a command to download an export.
func (m Model) downloadExportCmd(release *exports.Release) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Create a new temp directory for this download
		tempDir := filepath.Join(os.TempDir(), "rotector-exports")

		// Download to temp directory
		err := m.downloader.DownloadExport(ctx, release, tempDir)
		if err != nil {
			return ExportDownloadCompleteMsg{Error: err}
		}

		return ExportDownloadCompleteMsg{
			Error:   nil,
			TempDir: tempDir,
		}
	}
}
