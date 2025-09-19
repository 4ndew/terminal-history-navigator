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

		// Wrap long items and apply styling
		wrappedItem := m.wrapAndStyleItem(item, statusIndicator, i == selectedIndex)
		renderedItems = append(renderedItems, wrappedItem)
	}

	return strings.Join(renderedItems, "\n")
}

// wrapAndStyleItem wraps long text and applies styling
func (m Model) wrapAndStyleItem(item string, statusIndicator string, isSelected bool) string {
	// Calculate available width for text (account for indicators and padding)
	maxWidth := m.width - 12 // Account for status indicators, selection markers, padding
	if maxWidth < 20 {
		maxWidth = 20
	}

	// If item fits in one line, render normally
	if len(item) <= maxWidth {
		var styledItem string
		if isSelected {
			styledItem = selectedItemStyle.Render("► " + statusIndicator + item)
		} else {
			styledItem = normalItemStyle.Render("  " + statusIndicator + item)
		}
		return styledItem
	}

	// Item is too long - wrap it
	lines := wrapText(item, maxWidth)
	var wrappedLines []string

	for j, line := range lines {
		var prefix string
		var indicator string

		if j == 0 {
			// First line - normal selection marker and status
			if isSelected {
				prefix = "► "
			} else {
				prefix = "  "
			}
			indicator = statusIndicator
		} else {
			// Continuation lines - indent
			if isSelected {
				prefix = "  "
			} else {
				prefix = "  "
			}
			indicator = "  " // Same width as statusIndicator for alignment
		}

		var styledLine string
		if isSelected {
			styledLine = selectedItemStyle.Render(prefix + indicator + line)
		} else {
			styledLine = normalItemStyle.Render(prefix + indicator + line)
		}
		wrappedLines = append(wrappedLines, styledLine)
	}

	return strings.Join(wrappedLines, "\n")
}

// wrapText wraps text to specified width, preserving word boundaries where possible
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string

	for len(text) > 0 {
		if len(text) <= width {
			lines = append(lines, text)
			break
		}

		// Find the best break point
		breakPoint := width

		// Try to break at a space
		for i := width - 1; i >= width/2; i-- {
			if i < len(text) && text[i] == ' ' {
				breakPoint = i
				break
			}
		}

		// Take the line
		line := strings.TrimSpace(text[:breakPoint])
		lines = append(lines, line)

		// Continue with the rest
		text = strings.TrimSpace(text[breakPoint:])
	}

	return lines
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

	// Controls help
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
		return "enter: copy | t: history | /: search | ?: help | q: quit"
	default:
		// History mode controls
		return "enter: copy | t: templates | /: search | f: frequency | r: refresh | ?: help | q: quit"
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
  r           Refresh data from source files
  
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
