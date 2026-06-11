package ui

import (
	"fmt"
	"unicode/utf8"

	"github.com/4ndew/terminal-history-navigator/pkg/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleMouse processes mouse input (wheel scrolling).
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.showHelp {
		return m, nil
	}

	switch msg.Type {
	case tea.MouseWheelUp:
		for i := 0; i < 3; i++ {
			m.moveUp()
		}
	case tea.MouseWheelDown:
		for i := 0; i < 3; i++ {
			m.moveDown()
		}
	}

	return m, nil
}

// handleKeyPress processes keyboard input.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle help mode separately - any key closes help.
	if m.showHelp {
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		default:
			m.showHelp = false
			return m, nil
		}
	}

	switch m.mode {
	case SearchMode:
		return m.handleSearchKeys(msg)
	default:
		return m.handleNormalKeys(msg)
	}
}

// handleNormalKeys handles keys in normal (non-search) mode.
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

	case "f":
		// Toggle between frequency and chronological sort.
		if m.mode == HistoryMode {
			m.sortByFreq = !m.sortByFreq
			m.cursor = 0
			m.scrollOffset = 0
			m.loadCommands()
			if m.sortByFreq {
				m.setStatus("Sorted by frequency")
			} else {
				m.setStatus("Sorted chronologically (newest first)")
			}
		}
		return m, nil

	case "s":
		if m.mode == HistoryMode || m.mode == SearchMode {
			cmd := m.getCurrentItem()
			if cmd == "" {
				m.setError("Nothing selected")
				return m, nil
			}
			t, err := m.templateLoader.Add(cmd)
			if err != nil {
				m.setError(fmt.Sprintf("Failed to save template: %v", err))
				return m, nil
			}
			updated, _ := m.templateLoader.Load()
			m.templates = updated
			m.setStatus(fmt.Sprintf("Saved as template: %s [%s]", t.Name, t.Category))
		}
		return m, nil

	case "d":
		if m.mode == TemplatesMode {
			if len(m.templates) == 0 || m.cursor >= len(m.templates) {
				m.setError("Nothing selected")
				return m, nil
			}
			cmd := m.templates[m.cursor].Command
			name := m.templates[m.cursor].Name
			if err := m.templateLoader.Delete(cmd); err != nil {
				m.setError(fmt.Sprintf("Failed to delete template: %v", err))
				return m, nil
			}
			updated, _ := m.templateLoader.Load()
			m.templates = updated
			if m.cursor >= len(m.templates) && m.cursor > 0 {
				m.cursor--
			}
			m.setStatus(fmt.Sprintf("Deleted template: %s", name))
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

// handleSearchKeys handles keys in search mode.
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
			// Remove the last rune, not the last byte (UTF-8 safe).
			runes := []rune(m.searchQuery)
			m.setSearchQuery(string(runes[:len(runes)-1]))
		}
		return m, nil
	}

	// Regular character input: KeyRunes covers any printable input
	// including non-ASCII and pasted text; space arrives as KeySpace.
	switch msg.Type {
	case tea.KeyRunes:
		m.setSearchQuery(m.searchQuery + string(msg.Runes))
	case tea.KeySpace:
		m.setSearchQuery(m.searchQuery + " ")
	}

	return m, nil
}

// handleSelectItem handles selecting/copying the current item.
func (m Model) handleSelectItem() (tea.Model, tea.Cmd) {
	selectedText := m.getCurrentItem()
	if selectedText == "" {
		m.setError("No item selected")
		return m, nil
	}

	// Copy to clipboard (original text, multiline preserved).
	err := clipboard.Copy(selectedText)
	if err != nil {
		m.setError(fmt.Sprintf("Failed to copy: %v", err))
		return m, nil
	}

	m.setStatus(fmt.Sprintf("Copied: %s", truncateString(displayText(selectedText), 50)))

	return m, nil
}

// truncateString truncates a string to maxLen runes with ellipsis (UTF-8 safe).
func truncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen-3]) + "..."
}