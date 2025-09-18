package ui

import (
	"github.com/4ndew/terminal-history-navigator/internal/config"
	"github.com/4ndew/terminal-history-navigator/internal/history"
	"github.com/4ndew/terminal-history-navigator/internal/storage"
	"github.com/4ndew/terminal-history-navigator/internal/templates"
	tea "github.com/charmbracelet/bubbletea"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	HistoryMode ViewMode = iota
	TemplatesMode
	SearchMode
)

// Model represents the TUI application state
type Model struct {
	// Data
	storage   storage.Storage
	templates []templates.Template
	config    *config.Config

	// Current state
	commands     []history.Command
	filteredCmds []history.Command
	mode         ViewMode
	cursor       int
	searchQuery  string

	// UI state
	width    int
	height   int
	showHelp bool

	// Status messages
	statusMsg string
	errorMsg  string
}

// NewModel creates a new TUI model
func NewModel(store storage.Storage, templateList []templates.Template, cfg *config.Config) Model {
	model := Model{
		storage:   store,
		templates: templateList,
		config:    cfg,
		mode:      HistoryMode,
		cursor:    0,
		width:     80,
		height:    24,
	}

	// Load initial commands
	model.loadCommands()

	return model
}

// Init initializes the model (required by bubbletea)
func (m Model) Init() tea.Cmd {
	return nil
}

// loadCommands loads commands based on current mode and filters
func (m *Model) loadCommands() {
	switch m.mode {
	case HistoryMode:
		if m.searchQuery != "" {
			m.filteredCmds = m.storage.Search(m.searchQuery)
		} else {
			m.filteredCmds = m.storage.GetRecent(m.config.UI.MaxItems)
		}
	case TemplatesMode:
		// Templates are handled separately, clear filtered commands
		m.filteredCmds = []history.Command{}
	case SearchMode:
		// Search mode uses the same data as history mode
		m.filteredCmds = m.storage.Search(m.searchQuery)
	}

	// Reset cursor if it's out of bounds
	if m.cursor >= len(m.filteredCmds) {
		m.cursor = 0
	}
}

// getCurrentItem returns the currently selected item text
func (m *Model) getCurrentItem() string {
	switch m.mode {
	case HistoryMode, SearchMode:
		if len(m.filteredCmds) == 0 || m.cursor >= len(m.filteredCmds) {
			return ""
		}
		return m.filteredCmds[m.cursor].Text

	case TemplatesMode:
		if len(m.templates) == 0 || m.cursor >= len(m.templates) {
			return ""
		}
		return m.templates[m.cursor].Command
	}

	return ""
}

// moveUp moves the cursor up
func (m *Model) moveUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

// moveDown moves the cursor down
func (m *Model) moveDown() {
	maxItems := 0
	switch m.mode {
	case HistoryMode, SearchMode:
		maxItems = len(m.filteredCmds)
	case TemplatesMode:
		maxItems = len(m.templates)
	}

	if m.cursor < maxItems-1 {
		m.cursor++
	}
}

// setSearchQuery updates the search query and reloads commands
func (m *Model) setSearchQuery(query string) {
	m.searchQuery = query
	m.loadCommands()
}

// switchToHistoryMode switches to history view mode
func (m *Model) switchToHistoryMode() {
	m.mode = HistoryMode
	m.cursor = 0
	m.searchQuery = ""
	m.loadCommands()
	m.statusMsg = "History mode"
}

// switchToTemplatesMode switches to templates view mode
func (m *Model) switchToTemplatesMode() {
	m.mode = TemplatesMode
	m.cursor = 0
	m.statusMsg = "Templates mode"
}

// switchToSearchMode switches to search mode
func (m *Model) switchToSearchMode() {
	m.mode = SearchMode
	m.cursor = 0
	m.statusMsg = "Search mode - type to search"
}

// exitSearchMode exits search mode and returns to history
func (m *Model) exitSearchMode() {
	if m.mode == SearchMode {
		m.searchQuery = ""
		m.switchToHistoryMode()
	}
}

// setStatus sets a status message
func (m *Model) setStatus(msg string) {
	m.statusMsg = msg
	m.errorMsg = ""
}

// setError sets an error message
func (m *Model) setError(msg string) {
	m.errorMsg = msg
	m.statusMsg = ""
}

// clearMessages clears status and error messages
func (m *Model) clearMessages() {
	m.statusMsg = ""
	m.errorMsg = ""
}

// getVisibleItems returns the items that should be visible on screen
func (m *Model) getVisibleItems() ([]string, int) {
	var items []string
	var selectedIndex int

	switch m.mode {
	case HistoryMode, SearchMode:
		for _, cmd := range m.filteredCmds {
			items = append(items, cmd.Text)
		}
		selectedIndex = m.cursor

	case TemplatesMode:
		for _, template := range m.templates {
			// Format: "Name - Command (Description)"
			item := template.Name + " - " + template.Command
			if template.Description != "" {
				item += " (" + template.Description + ")"
			}
			items = append(items, item)
		}
		selectedIndex = m.cursor
	}

	return items, selectedIndex
}

// getItemCount returns the total number of items in current mode
func (m *Model) getItemCount() int {
	switch m.mode {
	case HistoryMode, SearchMode:
		return len(m.filteredCmds)
	case TemplatesMode:
		return len(m.templates)
	}
	return 0
}

// resize updates the model dimensions
func (m *Model) resize(width, height int) {
	m.width = width
	m.height = height
}
