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
			Bold(true)

	// Item styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(primaryColor).
				Padding(0, 1)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

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

	// Header - always show
	sections = append(sections, m.renderHeader())
	sections = append(sections, "") // Empty line for separation

	// Main content
	sections = append(sections, m.renderMainContent())

	// Footer
	sections = append(sections, "") // Empty line before footer
	sections = append(sections, m.renderFooter())

	return strings.Join(sections, "\n")
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

	// Calculate visible range with proper scrolling
	maxVisible := m.height - 6 // Account for header, separators, footer
	if maxVisible < 5 {
		maxVisible = 5
	}

	start := 0
	end := len(items)

	// If we have more items than can fit, calculate scroll window
	if len(items) > maxVisible {
		// Simple scrolling logic: keep selected item in view
		if selectedIndex < maxVisible {
			// If selection is in the first screen, start from 0
			start = 0
		} else {
			// Otherwise, scroll to show selected item
			start = selectedIndex - maxVisible + 1
			if start < 0 {
				start = 0
			}
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
		isSelected := (i == selectedIndex)

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

		// Render item
		renderedItem := m.renderSingleItem(item, statusIndicator, isSelected)
		renderedItems = append(renderedItems, renderedItem)
	}

	return strings.Join(renderedItems, "\n")
}

// renderSingleItem renders a single item with proper wrapping
func (m Model) renderSingleItem(item string, statusIndicator string, isSelected bool) string {
	// Calculate available width
	maxWidth := m.width - 6 // Account for selection markers and padding
	if maxWidth < 20 {
		maxWidth = 20
	}

	// Prepare the full text with status indicator
	fullText := statusIndicator + item

	var prefix string
	if isSelected {
		prefix = "► "
	} else {
		prefix = "  "
	}

	// If it fits in one line
	if len(prefix+fullText) <= maxWidth {
		var styledItem string
		if isSelected {
			styledItem = selectedItemStyle.Render(prefix + fullText)
		} else {
			styledItem = normalItemStyle.Render(prefix + fullText)
		}
		return styledItem
	}

	// Need to wrap
	availableForText := maxWidth - len(prefix) - len(statusIndicator)
	if availableForText < 10 {
		availableForText = 10
	}

	lines := wrapText(item, availableForText)
	var wrappedLines []string

	for j, line := range lines {
		var linePrefix string
		var indicator string

		if j == 0 {
			// First line gets the selection marker and status
			linePrefix = prefix
			indicator = statusIndicator
		} else {
			// Continuation lines get padding
			linePrefix = "  "
			indicator = strings.Repeat(" ", len(statusIndicator))
		}

		var styledLine string
		if isSelected {
			styledLine = selectedItemStyle.Render(linePrefix + indicator + line)
		} else {
			styledLine = normalItemStyle.Render(linePrefix + indicator + line)
		}
		wrappedLines = append(wrappedLines, styledLine)
	}

	return strings.Join(wrappedLines, "\n")
}

// wrapText wraps text to specified width
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	remaining := text

	for len(remaining) > 0 {
		if len(remaining) <= width {
			lines = append(lines, remaining)
			break
		}

		// Find best break point
		breakPoint := width
		for i := width - 1; i >= width/2 && i > 0; i-- {
			if i < len(remaining) && remaining[i] == ' ' {
				breakPoint = i
				break
			}
		}

		// Take the line and continue
		line := strings.TrimSpace(remaining[:breakPoint])
		lines = append(lines, line)
		remaining = strings.TrimSpace(remaining[breakPoint:])
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

	// Join sections and wrap if necessary
	footer := strings.Join(sections, " | ")
	return m.wrapFooter(footer)
}

// wrapFooter wraps the footer text if it exceeds screen width
func (m Model) wrapFooter(footer string) string {
	maxWidth := m.width - 4
	if maxWidth < 20 {
		maxWidth = 20
	}

	if len(footer) <= maxWidth {
		return footer
	}

	// Split and wrap footer
	parts := strings.Split(footer, " | ")
	var lines []string
	var currentLine string

	for i, part := range parts {
		testLine := currentLine
		if testLine != "" {
			testLine += " | "
		}
		testLine += part

		if len(testLine) <= maxWidth {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = part
		}

		if i == len(parts)-1 && currentLine != "" {
			lines = append(lines, currentLine)
		}
	}

	return strings.Join(lines, "\n")
}

// getControlsHelp returns context-appropriate control hints
func (m Model) getControlsHelp() string {
	switch m.mode {
	case SearchMode:
		return "esc: exit | enter: copy | ↑↓: navigate"
	case TemplatesMode:
		return "enter: copy | t: history | /: search | ?: help | q: quit"
	default:
		return "enter: copy | t: templates | /: search | f: frequency | ?: help | q: quit"
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
