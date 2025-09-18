package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	// Colors
	primaryColor = lipgloss.Color("#00D4AA")
	accentColor  = lipgloss.Color("#F59E0B")
	mutedColor   = lipgloss.Color("#6B7280")
	errorColor   = lipgloss.Color("#EF4444")
	successColor = lipgloss.Color("#10B981")

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	// Item styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(primaryColor).
				Padding(0, 1)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB")).
			Padding(0, 1)

	// Footer styles
	footerStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Search styles
	searchStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// Help styles
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Border(lipgloss.RoundedBorder()).
			Padding(1).
			Margin(1)
)

// View renders the TUI interface
func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	var sections []string

	// Header
	sections = append(sections, m.renderHeader())

	// Main content
	sections = append(sections, m.renderMainContent())

	// Footer
	sections = append(sections, m.renderFooter())

	return baseStyle.Render(strings.Join(sections, "\n"))
}

// renderHeader renders the application header
func (m Model) renderHeader() string {
	title := "Terminal History Navigator"

	var modeStr string
	switch m.mode {
	case HistoryMode:
		modeStr = "History"
	case TemplatesMode:
		modeStr = "Templates"
	case SearchMode:
		modeStr = fmt.Sprintf("Search: %s", m.searchQuery)
	}

	header := headerStyle.Render(title) + " - " + searchStyle.Render(modeStr)
	return header
}

// renderMainContent renders the main content area
func (m Model) renderMainContent() string {
	items, selectedIndex := m.getVisibleItems()

	if len(items) == 0 {
		return m.renderEmptyState()
	}

	var renderedItems []string

	// Calculate visible range (simple scrolling)
	maxVisible := m.height - 6 // Account for header, footer, padding
	if maxVisible < 1 {
		maxVisible = 10
	}

	start := 0
	end := len(items)

	// If there are more items than can fit, center the selection
	if len(items) > maxVisible {
		start = selectedIndex - maxVisible/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(items) {
			end = len(items)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
	}

	// Render visible items
	for i := start; i < end; i++ {
		item := items[i]

		// Add status indicator for commands with exit codes
		statusIndicator := ""
		if m.mode == HistoryMode || m.mode == SearchMode {
			if i < len(m.filteredCmds) {
				cmd := m.filteredCmds[i]
				if cmd.HasExit {
					if cmd.ExitCode == 0 {
						statusIndicator = lipgloss.NewStyle().Foreground(successColor).Render("✓ ")
					} else {
						statusIndicator = lipgloss.NewStyle().Foreground(errorColor).Render("✗ ")
					}
				}
			}
		}

		// Truncate long items (account for status indicator)
		maxWidth := m.width - 12
		if maxWidth < 20 {
			maxWidth = 20
		}
		if len(item) > maxWidth {
			item = item[:maxWidth-3] + "..."
		}

		// Apply styling based on selection
		var styledItem string
		if i == selectedIndex {
			styledItem = selectedItemStyle.Render("► " + statusIndicator + item)
		} else {
			styledItem = normalItemStyle.Render("  " + statusIndicator + item)
		}

		renderedItems = append(renderedItems, styledItem)
	}

	return strings.Join(renderedItems, "\n")
}

// renderEmptyState renders the empty state message
func (m Model) renderEmptyState() string {
	var message string

	switch m.mode {
	case HistoryMode:
		message = "No command history found"
	case TemplatesMode:
		message = "No templates available"
	case SearchMode:
		if m.searchQuery == "" {
			message = "Start typing to search..."
		} else {
			message = fmt.Sprintf("No results for '%s'", m.searchQuery)
		}
	}

	return lipgloss.NewStyle().Foreground(mutedColor).Render(message)
}

// renderFooter renders the footer with status and controls
func (m Model) renderFooter() string {
	var sections []string

	// Status or error message
	if m.errorMsg != "" {
		sections = append(sections, errorStyle.Render("Error: "+m.errorMsg))
	} else if m.statusMsg != "" {
		sections = append(sections, statusStyle.Render(m.statusMsg))
	}

	// Item count and position info
	itemCount := m.getItemCount()
	if itemCount > 0 {
		position := fmt.Sprintf("%d/%d", m.cursor+1, itemCount)

		// Add sorting info
		var sortInfo string
		if m.mode == HistoryMode {
			if m.statusMsg == "Sorted by frequency" {
				sortInfo = " (by frequency)"
			} else {
				sortInfo = " (newest first)"
			}
		}

		sections = append(sections, lipgloss.NewStyle().Foreground(mutedColor).Render(position+sortInfo))
	}

	// Controls help (can be multi-line now)
	controls := m.getControlsHelp()
	sections = append(sections, footerStyle.Render(controls))

	return strings.Join(sections, " | ")
}

// getControlsHelp returns context-appropriate control hints
func (m Model) getControlsHelp() string {
	switch m.mode {
	case SearchMode:
		return "esc: exit search | enter: select | ↑↓: navigate"
	case TemplatesMode:
		return "enter: copy | e: edit | t: history | /: search\n?: help | q: quit"
	default:
		// Split history mode controls into two lines
		line1 := "enter: copy | t: templates | /: search | f: frequency"
		line2 := "s: successful | x: failed | r: refresh | ?: help | q: quit"
		return line1 + "\n" + line2
	}
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	helpText := `Terminal History Navigator - Help

NAVIGATION:
  ↑/k         Move up
  ↓/j         Move down
  enter       Copy selected item to clipboard
  
MODES:
  h           Switch to history mode
  t           Toggle templates mode
  /           Start search
  f           Sort by frequency (history mode)
  r           Refresh data
  e           Edit templates (templates mode)
  
SEARCH:
  /           Enter search mode
  esc         Exit search mode
  backspace   Delete search character
  
OTHER:
  ?           Toggle this help
  esc         Clear messages / close help
  q/ctrl+c    Quit application

CONFIGURATION:
  Config: ~/.config/history-nav/config.yaml
  Templates: ~/.config/history-nav/templates.yaml

Press any key to close help...`

	return helpStyle.Render(helpText)
}
