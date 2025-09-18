package ui

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/4ndew/terminal-history-navigator/internal/history"
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
		// Refresh commands
		m.loadCommands()
		m.setStatus("Refreshed")
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

	case "e":
		// Edit templates (only in templates mode)
		if m.mode == TemplatesMode {
			return m.handleEditTemplates()
		}
		return m, nil

	case "esc":
		m.clearMessages()
		m.showHelp = false
		return m, nil

	case "s":
		// Show only successful commands (exit code 0)
		if m.mode == HistoryMode {
			allCommands := m.storage.GetRecent(m.config.UI.MaxItems)
			var successfulCmds []history.Command
			for _, cmd := range allCommands {
				if !cmd.HasExit || cmd.ExitCode == 0 {
					successfulCmds = append(successfulCmds, cmd)
				}
			}
			m.filteredCmds = successfulCmds
			m.cursor = 0
			m.setStatus("Showing only successful commands")
		}
		return m, nil

	case "x":
		// Show only failed commands (exit code != 0)
		if m.mode == HistoryMode {
			allCommands := m.storage.GetRecent(m.config.UI.MaxItems)
			var failedCmds []history.Command
			for _, cmd := range allCommands {
				if cmd.HasExit && cmd.ExitCode != 0 {
					failedCmds = append(failedCmds, cmd)
				}
			}
			m.filteredCmds = failedCmds
			m.cursor = 0
			m.setStatus("Showing only failed commands")
		}
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

// handleEditTemplates opens the templates file in an editor
func (m Model) handleEditTemplates() (tea.Model, tea.Cmd) {
	// Get templates file path from config
	templatesPath := m.config.TemplatesPath

	// Try to open in default editor
	editors := []string{
		os.Getenv("EDITOR"),
		"nano",
		"vim",
		"code",
		"open", // macOS default
	}

	for _, editor := range editors {
		if editor == "" {
			continue
		}

		// For "open" on macOS, use -t flag for text editor
		var cmd *exec.Cmd
		if editor == "open" {
			cmd = exec.Command("open", "-t", templatesPath)
		} else {
			cmd = exec.Command(editor, templatesPath)
		}

		if err := cmd.Start(); err == nil {
			m.setStatus("Opening templates in " + editor + ". Restart app to see changes.")
			return m, nil
		}
	}

	// If no editor worked, show path
	m.setStatus("Edit templates: " + templatesPath)
	return m, nil
}
