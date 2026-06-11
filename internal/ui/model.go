package ui

import (
	"fmt"
	"strings"

	"github.com/4ndew/terminal-history-navigator/internal/config"
	"github.com/4ndew/terminal-history-navigator/internal/history"
	"github.com/4ndew/terminal-history-navigator/internal/storage"
	"github.com/4ndew/terminal-history-navigator/internal/templates"
	tea "github.com/charmbracelet/bubbletea"
)

// ViewMode represents the current view mode.
type ViewMode int

const (
	HistoryMode ViewMode = iota
	TemplatesMode
	SearchMode
)

// Model represents the TUI application state.
type Model struct {
	// Data
	storage        storage.Storage
	templates      []templates.Template
	templateLoader *templates.Loader
	config         *config.Config

	// Current state
	commands     []history.Command // All available commands
	filteredCmds []history.Command // Filtered commands for display
	mode         ViewMode
	cursor       int
	scrollOffset int
	searchQuery  string
	sortByFreq   bool // History mode: sort by real frequency instead of recency

	// UI state
	width    int
	height   int
	showHelp bool

	// Status messages
	statusMsg string
	errorMsg  string
}

// NewModel creates a new TUI model.
func NewModel(store storage.Storage, templateList []templates.Template, loader *templates.Loader, cfg *config.Config) Model {
	model := Model{
		storage:        store,
		templates:      templateList,
		templateLoader: loader,
		config:         cfg,
		mode:           HistoryMode,
		cursor:         0,
		width:          80,
		height:         24,
	}

	model.loadCommands()
	return model
}

// Init initializes the model (required by bubbletea).
func (m Model) Init() tea.Cmd {
	return nil
}

// loadCommands loads commands based on current mode, sort order and filters.
func (m *Model) loadCommands() {
	m.commands = m.storage.GetAll()

	switch m.mode {
	case HistoryMode:
		if m.searchQuery != "" {
			m.filteredCmds = m.storage.Search(m.searchQuery)
		} else if m.sortByFreq {
			cmds := m.storage.GetByFrequency()
			if len(cmds) > m.config.UI.MaxItems {
				cmds = cmds[:m.config.UI.MaxItems]
			}
			m.filteredCmds = cmds
		} else {
			m.filteredCmds = m.storage.GetRecent(m.config.UI.MaxItems)
		}
	case TemplatesMode:
		m.filteredCmds = []history.Command{}
	case SearchMode:
		m.filteredCmds = m.storage.Search(m.searchQuery)
	}

	// Reset cursor if it's out of bounds.
	if m.mode != TemplatesMode && m.cursor >= len(m.filteredCmds) {
		m.cursor = 0
		m.scrollOffset = 0
	}
	if m.mode == TemplatesMode && m.cursor >= len(m.templates) {
		m.cursor = 0
		m.scrollOffset = 0
	}
}

// getCurrentItem returns the currently selected item text (original,
// with real newlines for multiline commands — this is what gets copied).
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

// displayText converts a command to a single-line representation for the list.
func displayText(s string) string {
	return strings.ReplaceAll(s, "\n", " ⏎ ")
}

// moveUp moves the cursor up.
func (m *Model) moveUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

// moveDown moves the cursor down.
func (m *Model) moveDown() {
	maxItems := m.getItemCount()
	if m.cursor < maxItems-1 {
		m.cursor++
	}
}

func (m *Model) adjustScrollOffset(itemHeights []int, visibleLines int) {
	if len(itemHeights) == 0 {
		m.scrollOffset = 0
		return
	}

	half := visibleLines / 2

	start := m.cursor
	accumulated := 0
	for i := m.cursor - 1; i >= 0; i-- {
		if accumulated+itemHeights[i] > half {
			break
		}
		accumulated += itemHeights[i]
		start = i
	}

	m.scrollOffset = start
}

// setSearchQuery updates the search query and reloads commands.
func (m *Model) setSearchQuery(query string) {
	m.searchQuery = query
	m.loadCommands()
}

// switchToHistoryMode switches to history view mode.
func (m *Model) switchToHistoryMode() {
	m.mode = HistoryMode
	m.cursor = 0
	m.scrollOffset = 0
	m.searchQuery = ""
	m.loadCommands()
	m.statusMsg = ""
}

// switchToTemplatesMode switches to templates view mode.
func (m *Model) switchToTemplatesMode() {
	m.mode = TemplatesMode
	m.cursor = 0
	m.scrollOffset = 0
	m.statusMsg = ""
}

// switchToSearchMode switches to search mode.
func (m *Model) switchToSearchMode() {
	m.mode = SearchMode
	m.cursor = 0
	m.scrollOffset = 0
	m.statusMsg = ""
}

// exitSearchMode exits search mode and returns to history.
func (m *Model) exitSearchMode() {
	if m.mode == SearchMode {
		m.searchQuery = ""
		m.switchToHistoryMode()
	}
}

// setStatus sets a status message.
func (m *Model) setStatus(msg string) {
	m.statusMsg = msg
	m.errorMsg = ""
}

// setError sets an error message.
func (m *Model) setError(msg string) {
	m.errorMsg = msg
	m.statusMsg = ""
}

// clearMessages clears status and error messages.
func (m *Model) clearMessages() {
	m.statusMsg = ""
	m.errorMsg = ""
}

// getVisibleItems returns the items that should be visible on screen.
func (m *Model) getVisibleItems() ([]string, int) {
	var items []string
	var selectedIndex int

	switch m.mode {
	case HistoryMode, SearchMode:
		for _, cmd := range m.filteredCmds {
			item := displayText(cmd.Text)
			if m.sortByFreq && m.mode == HistoryMode && cmd.Count > 1 {
				item = fmt.Sprintf("[%dx] %s", cmd.Count, item)
			}
			items = append(items, item)
		}
		selectedIndex = m.cursor

	case TemplatesMode:
		for _, template := range m.templates {
			item := template.Name + " - " + displayText(template.Command)
			if template.Description != "" {
				item += " (" + template.Description + ")"
			}
			items = append(items, item)
		}
		selectedIndex = m.cursor
	}

	return items, selectedIndex
}

// getItemCount returns the total number of items in current mode.
func (m *Model) getItemCount() int {
	switch m.mode {
	case HistoryMode, SearchMode:
		return len(m.filteredCmds)
	case TemplatesMode:
		return len(m.templates)
	}
	return 0
}

// resize updates the model dimensions.
func (m *Model) resize(width, height int) {
	m.width = width
	m.height = height
}