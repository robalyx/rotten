//nolint:gochecknoglobals
package tui

import "github.com/charmbracelet/lipgloss"

var (
	// boxStyle defines the main container box style with rounded borders.
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")). // Cyan blue
			Padding(1).
			Width(60)

	// programTitleStyle defines the main program title appearance.
	programTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("33")). // Deep blue
				Bold(true).
				MarginBottom(0).
				Align(lipgloss.Center).
				Width(58)

	// subtitleStyle defines the subtitle appearance below the program title.
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")). // Gray
			Align(lipgloss.Center).
			MarginBottom(1).
			Width(58)

	// titleStyle defines the section title appearance.
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")). // Cyan blue
			Bold(true).
			MarginBottom(1).
			Align(lipgloss.Center)

	// selectedStyle defines the appearance of selected menu items.
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("45")). // Bright cyan
			Bold(true)

	// optionStyle defines the appearance of unselected menu items.
	optionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")) // Bright white

	// successStyle defines the appearance of success messages.
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("35")). // Green
			Bold(true)

	// failureStyle defines the appearance of error messages.
	failureStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red
			Bold(true)

	// inputStyle defines the appearance of user input text.
	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("51")). // Bright cyan
			Bold(true)

	// helpStyle defines the appearance of help text and instructions.
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")). // Gray
			Italic(true).
			MarginTop(1).
			Align(lipgloss.Center)

	// textFieldStyle defines the appearance of input text fields.
	textFieldStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("39")). // Cyan blue
			Padding(0, 1).
			Width(20)

	// confidenceStyle defines the appearance of confidence values.
	confidenceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // Orange
			Bold(true)

	// reasonBoxStyle defines the appearance of reason text boxes.
	reasonBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")). // Cyan blue
			Padding(0, 1).
			Width(50)
)
