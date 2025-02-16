package tui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/robalyx/rotten/internal/exports"
)

// View renders the current UI state as a string.
func (m Model) View() string {
	header := m.renderHeader()

	if m.err != nil {
		return m.renderError(header)
	}

	switch m.state {
	case StateCheckType:
		return m.renderCheckTypeView(header)
	case StateDirectory:
		return m.renderDirectoryView(header)
	case StateStorageType:
		return m.renderStorageTypeView(header)
	case StateIDInput:
		return m.renderIDInputView(header)
	case StateUserGroupResult:
		return m.renderResultView(header)
	case StateFriendsResult:
		return m.renderFriendsResultView(header)
	case StateExportDownload:
		return m.renderExportDownloadView(header)
	default:
		return boxStyle.Render(header + "\n\nUnknown state")
	}
}

// renderHeader renders the program title and subtitle.
func (m Model) renderHeader() string {
	title := programTitleStyle.Render("Rotten")
	subtitle := subtitleStyle.Render("A tool to check exports from Rotector")
	return fmt.Sprintf("%s\n%s", title, subtitle)
}

// renderError renders error messages with instructions.
func (m Model) renderError(header string) string {
	content := fmt.Sprintf("%s\n\n%s\n%s",
		header,
		failureStyle.Render(fmt.Sprintf("Error: %v", m.err)),
		helpStyle.Render("\nPress enter to restart or ctrl+c to quit"))
	return boxStyle.Render(content)
}

// renderCheckTypeView renders the check type selection menu.
func (m Model) renderCheckTypeView(header string) string {
	options := []string{"User", "Group", "Friends"}
	optionsText := ""
	for i, option := range options {
		if i == m.checkTypeSelected {
			optionsText += selectedStyle.Render("> " + option)
		} else {
			optionsText += optionStyle.Render("  " + option)
		}
		optionsText += "\n"
	}

	content := fmt.Sprintf("%s\n\n%s\n\n%s%s",
		header,
		titleStyle.Render("What would you like to check?"),
		optionsText,
		fmt.Sprintf("%s\n%s",
			helpStyle.Render("Use arrow keys to select and enter to confirm"),
			helpStyle.Render("Press 'r' to start over or ctrl+c to quit")))
	return boxStyle.Render(content)
}

// renderDirectoryView renders the directory selection menu.
func (m Model) renderDirectoryView(header string) string {
	optionsText := ""

	options := append([]string{"Download Official Export"}, m.directories...)
	for i, option := range options {
		if i == m.selected {
			optionsText += selectedStyle.Render("> " + option)
		} else {
			optionsText += optionStyle.Render("  " + option)
		}
		optionsText += "\n"
	}

	content := fmt.Sprintf("%s\n\n%s\n\n%s\n%s\n%s",
		header,
		titleStyle.Render("Select a directory:"),
		optionsText,
		helpStyle.Render("Use arrow keys to select and enter to confirm"),
		helpStyle.Render("Press 'r' to start over or ctrl+c to quit"))
	return boxStyle.Render(content)
}

// renderStorageTypeView renders the storage type selection menu.
func (m Model) renderStorageTypeView(header string) string {
	options := []string{"SQLite", "Binary", "CSV"}
	optionsText := ""
	for i, option := range options {
		if i == m.storageTypeSelected {
			optionsText += selectedStyle.Render("> " + option)
		} else {
			optionsText += optionStyle.Render("  " + option)
		}
		optionsText += "\n"
	}

	content := fmt.Sprintf("%s\n\n%s\n%s\n\n%s\n%s\n%s",
		header,
		titleStyle.Render("Select storage type:"),
		optionStyle.Render("(SQLite is recommended if you're unsure)"),
		optionsText,
		helpStyle.Render("Use arrow keys to select and enter to confirm"),
		helpStyle.Render("Press 'r' to start over or ctrl+c to quit"))
	return boxStyle.Render(content)
}

// renderIDInputView renders the ID input field with export info.
func (m Model) renderIDInputView(header string) string {
	textField := textFieldStyle.Render(m.id)
	if len(m.id) == 0 {
		placeholderStyle := textFieldStyle.
			BorderForeground(lipgloss.Color("39")).
			Foreground(lipgloss.Color("240"))
		textField = placeholderStyle.Render("Enter ID...")
	}

	exportInfo := fmt.Sprintf("Export Info:\n"+
		"• Hash Type: %s\n"+
		"• Storage: %s\n"+
		"• Available Hashes: %d\n"+
		"• Engine Version: %s\n"+
		"• Export Version: %s\n"+
		"• Description: %s\n"+
		"• Salt: %s\n",
		m.config.HashType,
		m.storageType,
		m.hashCount,
		m.config.EngineVersion,
		m.config.ExportVersion,
		m.config.Description,
		m.config.Salt)

	var statusText string
	var helpText string
	if m.checking {
		statusText = "\n" + successStyle.Render("Checking ID...")
		helpText = helpStyle.Render("Please wait...")
	} else {
		helpText = fmt.Sprintf("%s\n%s",
			helpStyle.Render("Press enter when done"),
			helpStyle.Render("Press 'r' to start over or ctrl+c to quit"))
	}

	content := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s%s\n\n%s",
		header,
		titleStyle.Render("Enter numeric ID to check:"),
		textField,
		optionStyle.Render(exportInfo),
		statusText,
		helpText)
	return boxStyle.Render(content)
}

// renderResultView renders the check results with status and reason.
func (m Model) renderResultView(header string) string {
	var resultText string
	if m.result {
		resultText = successStyle.Render("✓ FOUND")
	} else {
		resultText = failureStyle.Render("✗ NOT FOUND")
	}

	checkTypeStr := string(m.checkType)
	checkTypeStr = strings.ToUpper(checkTypeStr[:1]) + checkTypeStr[1:]

	var details string
	if m.result {
		formattedReason := strings.ReplaceAll(m.reason, "; ", "\n\n")
		details = fmt.Sprintf("\nStatus: %s\nConfidence: %s\nReason: %s",
			successStyle.Render(m.status),
			confidenceStyle.Render(fmt.Sprintf("%.2f", m.confidence)),
			reasonBoxStyle.Render(formattedReason))
	}

	content := fmt.Sprintf("%s\n\n%s\n\n%s ID %s was %s in the export%s\n\n%s\n%s",
		header,
		titleStyle.Render("Result:"),
		checkTypeStr,
		inputStyle.Render(m.id),
		resultText,
		details,
		helpStyle.Render("Press enter to check another ID"),
		helpStyle.Render("Press 'r' to start over or ctrl+c to quit"))
	return boxStyle.Render(content)
}

// renderFriendsResultView renders the friend check results.
func (m Model) renderFriendsResultView(header string) string {
	if m.checking {
		return boxStyle.Render(fmt.Sprintf("%s\n\n%s\n\n%s",
			header,
			titleStyle.Render("Checking Friends..."),
			helpStyle.Render("Please wait...")))
	}

	content := fmt.Sprintf("%s\n\n%s\n\n",
		header,
		titleStyle.Render("Friend Check Results:"))

	if len(m.friendResults) == 0 {
		content += optionStyle.Render("No friends found")
	} else {
		// Show scroll indicator if needed
		if m.friendsScrollPos > 0 {
			content += selectedStyle.Render("↑ More above") + "\n\n"
		}

		// Show current friend
		if result := m.friendResults[m.friendsScrollPos]; result.Found {
			formattedReason := strings.ReplaceAll(result.Reason, "; ", "\n\n")
			content += fmt.Sprintf("%s %s: %s\nConfidence: %s\n%s\n\n",
				failureStyle.Render("✗"),
				inputStyle.Render(strconv.FormatUint(result.ID, 10)),
				successStyle.Render(result.Status),
				confidenceStyle.Render(fmt.Sprintf("%.2f", result.Confidence)),
				reasonBoxStyle.Render(formattedReason))
		}

		// Show scroll indicator if needed
		if m.friendsScrollPos < len(m.friendResults)-1 {
			content += selectedStyle.Render("↓ More below") + "\n\n"
		}

		content += fmt.Sprintf("\n%s flagged friends found out of %d total friends",
			failureStyle.Render(strconv.Itoa(m.flaggedFriendCount)),
			m.totalFriendCount)
	}

	content += fmt.Sprintf("\n\n%s\n%s\n%s",
		helpStyle.Render("Use up/down arrows to scroll"),
		helpStyle.Render("Press enter to check another ID"),
		helpStyle.Render("Press 'r' to start over or ctrl+c to quit"))

	return boxStyle.Render(content)
}

// renderExportDownloadView renders the export download interface.
func (m Model) renderExportDownloadView(header string) string {
	if m.downloadError != nil {
		if errors.Is(m.downloadError, exports.ErrNewerVersionAvailable) {
			return boxStyle.Render(fmt.Sprintf("%s\n\n%s\n\n%s\n%s",
				header,
				failureStyle.Render("A newer version of Rotten is available!"),
				optionStyle.Render("Please visit https://github.com/robalyx/rotten/releases to update."),
				helpStyle.Render("Press 'r' to start over or ctrl+c to quit")))
		}
		return boxStyle.Render(fmt.Sprintf("%s\n\n%s\n%s",
			header,
			failureStyle.Render(fmt.Sprintf("Download failed: %v", m.downloadError)),
			helpStyle.Render("Press 'r' to start over or ctrl+c to quit")))
	}

	if len(m.availableExports) == 0 {
		return boxStyle.Render(fmt.Sprintf("%s\n\n%s\n%s",
			header,
			optionStyle.Render("No compatible exports available"),
			helpStyle.Render("Press 'r' to start over or ctrl+c to quit")))
	}

	var content string
	content = fmt.Sprintf("%s\n\n%s\n\n",
		header,
		titleStyle.Render("Available Exports"))

	for i, release := range m.availableExports {
		switch {
		case m.downloading:
			content += optionStyle.Render("  " + release.Name)
		case i == m.selectedExport:
			content += selectedStyle.Render("> " + release.Name)
		default:
			content += optionStyle.Render("  " + release.Name)
		}
		content += "\n"
	}

	if m.downloading {
		content += "\n" + successStyle.Render("Downloading...")
	}

	content += fmt.Sprintf("\n\n%s\n%s",
		helpStyle.Render("Use arrow keys to select and enter to download"),
		helpStyle.Render("Press 'r' to start over or ctrl+c to quit"))

	return boxStyle.Render(content)
}
