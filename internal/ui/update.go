package ui

import (
	"fmt"

	"github.com/4ndew/terminal-history-navigator/pkg/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case SearchMode:
		return m.handleSearchKeys(msg)
	default:
		return m.handleNormalKeys(msg)
	}
}

// handleNormalKeys handles keys in normal (non-search) mode
func (m Model) handleNormalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		m.moveUp()
		return m, nil

	case "down", "j":
		m.moveDown()
		return m, nil

	case "enter":
		return m.handleSelectItem()

	case "/":
		m.switchToSearchMode()
		return m, nil

	case "t":
		if m.mode == TemplatesMode {
			m.switchToHistoryMode()
		} else {
			m.switchToTemplatesMode()
		}
		return m, nil

	case "h":
		m.switchToHistoryMode()
		return m, nil

	case "r":
		// Refresh commands from source files
		err := m.refreshAllData()
		if err != nil {
			m.setError(fmt.Sprintf("Refresh failed: %v", err))
		} else {
			m.setStatus("Refreshed from source files")
		}
		return m, nil

	case "f":
		// Toggle between frequency and chronological sort
		if m.mode == HistoryMode {
			if m.statusMsg == "Sorted by frequency" {
				// Switch back to chronological
				m.filteredCmds = m.storage.GetRecent(m.config.UI.MaxItems)
				m.cursor = 0
				m.setStatus("Sorted chronologically (newest first)")
			} else {
				// Switch to frequency
				freqCmds := m.storage.GetByFrequency()
				if len(freqCmds) > m.config.UI.MaxItems {
					freqCmds = freqCmds[:m.config.UI.MaxItems]
				}
				m.filteredCmds = freqCmds
				m.cursor = 0
				m.setStatus("Sorted by frequency")
			}
		}
		return m, nil

	case "?":
		m.showHelp = !m.showHelp
		return m, nil

	case "esc":
		m.clearMessages()
		m.showHelp = false
		return m, nil
	}

	return m, nil
}

// handleSearchKeys handles keys in search mode
func (m Model) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.exitSearchMode()
		return m, nil

	case "enter":
		return m.handleSelectItem()

	case "up", "ctrl+p":
		m.moveUp()
		return m, nil

	case "down", "ctrl+n":
		m.moveDown()
		return m, nil

	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.setSearchQuery(m.searchQuery)
		}
		return m, nil

	default:
		// Handle regular character input
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
			m.setSearchQuery(m.searchQuery)
		}
		return m, nil
	}
}

// handleSelectItem handles selecting/copying the current item
func (m Model) handleSelectItem() (tea.Model, tea.Cmd) {
	selectedText := m.getCurrentItem()
	if selectedText == "" {
		m.setError("No item selected")
		return m, nil
	}

	// Copy to clipboard
	err := clipboard.Copy(selectedText)
	if err != nil {
		m.setError(fmt.Sprintf("Failed to copy: %v", err))
		return m, nil
	}

	// Show success message
	m.setStatus(fmt.Sprintf("Copied: %s", truncateString(selectedText, 50)))

	return m, nil
}

// truncateString truncates a string to maxLen characters with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
