package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaxron/roapi.go/pkg/api"
	"github.com/robalyx/rotten/internal/checker"
	"github.com/robalyx/rotten/internal/common"
	"github.com/robalyx/rotten/internal/config"
	"github.com/robalyx/rotten/internal/exports"
)

const OfficialExportDir = "Old Downloaded Export"

// Model handles the state and behavior of the TUI.
type Model struct {
	// API clients
	roAPI      *api.API
	downloader *exports.Downloader

	// Configuration
	config *config.Config

	// Core state
	state     State
	err       error
	checkType common.CheckType

	// Directory selection
	directories []string
	selected    int

	// Storage configuration
	storageType         common.StorageType
	storageTypeSelected int

	// Check type selection
	checkTypeSelected int

	// ID input and validation
	id        string
	validator *checker.Validator
	checker   checker.Checker
	hashCount uint64

	// Check results
	result     bool
	status     string
	reason     string
	confidence float64

	// Friends check specific
	friendResults      []FriendResult
	friendsScrollPos   int
	flaggedFriendCount int
	totalFriendCount   int

	// Export download specific
	availableExports []*exports.Release
	selectedExport   int
	downloadError    error
	downloading      bool
	checking         bool
}

// FriendResult represents the result of a friend check.
type FriendResult struct {
	ID         uint64
	Found      bool
	Status     string
	Reason     string
	Confidence float64
}

// NewModel creates a new Model instance.
func NewModel() *Model {
	validator := checker.NewValidator()

	// Only get exports from current directory
	currentDirs, err := validator.GetExportDirs(".")

	var dirs []string
	if err == nil && len(currentDirs) > 0 {
		dirs = append(dirs, currentDirs...)
	}

	return &Model{
		roAPI:       api.New(nil),
		state:       StateCheckType,
		validator:   validator,
		directories: dirs,
		err:         err,
		downloader:  exports.New("robalyx", "rotten"),
		downloading: false,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	// If we're starting in export download state, load exports
	if m.state == StateExportDownload {
		return m.loadExportsCmd()
	}
	return nil
}
