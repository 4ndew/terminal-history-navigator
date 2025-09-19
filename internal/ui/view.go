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

	// Item styles - менее яркий цвет выделения
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#4A5568")). // Менее яркий серо-синий
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

	// Header - always show in all modes
	sections = append(sections, m.renderHeader())
	sections = append(sections, "") // Empty line for separation

	// Main content
	sections = append(sections, m.renderMainContent())

	// Footer
	sections = append(sections, "") // Empty line before footer
	sections = append(sections, m.renderFooter())

	return strings.Join(sections, "\n")
}

// renderHeader renders the application header - always visible in all modes
func (m Model) renderHeader() string {
	title := headerStyle.Render("Terminal History Navigator")

	var modeStr string
	switch m.mode {
	case HistoryMode:
		modeStr = "History"
	case TemplatesMode:
		modeStr = "Templates"
	case SearchMode:
		if m.searchQuery == "" {
			modeStr = "Search"
		} else {
			modeStr = fmt.Sprintf("Search: %s", m.searchQuery)
		}
	}

	modeDisplay := searchStyle.Render(fmt.Sprintf("[%s]", modeStr))
	return title + " " + modeDisplay
}

// renderMainContent renders the main content area with improved scrolling for multiline items
func (m Model) renderMainContent() string {
	items, selectedIndex := m.getVisibleItems()

	if len(items) == 0 {
		return m.renderEmptyState()
	}

	// Calculate available space for items (subtract header, separators, footer)
	maxVisibleLines := m.height - 6 // Header(1) + separator(1) + separator(1) + footer(3)
	if maxVisibleLines < 3 {
		maxVisibleLines = 3
	}

	// Pre-calculate how many lines each item will take
	itemHeights := make([]int, len(items))
	totalLines := 0

	for i, item := range items {
		height := m.calculateItemHeight(item, i == selectedIndex)
		itemHeights[i] = height
		totalLines += height
	}

	// If all items fit, show them all
	if totalLines <= maxVisibleLines {
		return m.renderItemsRange(items, 0, len(items), selectedIndex, itemHeights)
	}

	// Calculate scroll window considering item heights
	start, end := m.calculateScrollWindowForMultiline(items, itemHeights, selectedIndex, maxVisibleLines)
	return m.renderItemsRange(items, start, end, selectedIndex, itemHeights)
}

// calculateItemHeight calculates how many lines an item will occupy
func (m Model) calculateItemHeight(item string, isSelected bool) int {
	maxWidth := m.width - 6 // Account for selection markers and padding
	if maxWidth < 20 {
		maxWidth = 20
	}

	var prefix string
	if isSelected {
		prefix = "► "
	} else {
		prefix = "  "
	}

	// Add status indicator space (approximate)
	statusIndicatorSpace := 2 // "✓ " or "✗ " or empty

	availableForText := maxWidth - len(prefix) - statusIndicatorSpace
	if availableForText < 10 {
		availableForText = 10
	}

	// If it fits in one line
	if len(item) <= availableForText {
		return 1
	}

	// Calculate wrapped lines
	lines := wrapText(item, availableForText)
	return len(lines)
}

// calculateScrollWindowForMultiline calculates scroll window considering multiline items
func (m Model) calculateScrollWindowForMultiline(items []string, itemHeights []int, selectedIndex, maxVisibleLines int) (int, int) {
	if selectedIndex < 0 || selectedIndex >= len(items) {
		return 0, len(items)
	}

	// Try different start positions to find one that fits selected item in view
	bestStart := 0
	bestEnd := len(items)

	// Start from selected item and work backwards
	for start := selectedIndex; start >= 0; start-- {
		currentLines := 0
		end := start

		// Count forward from start position
		for i := start; i < len(items) && currentLines < maxVisibleLines; i++ {
			if currentLines+itemHeights[i] <= maxVisibleLines {
				currentLines += itemHeights[i]
				end = i + 1
			} else {
				break
			}
		}

		// If selected item is visible in this window
		if selectedIndex >= start && selectedIndex < end {
			bestStart = start
			bestEnd = end

			// If we have room and selected item is not centered, continue looking
			selectedPosition := 0
			for i := start; i < selectedIndex; i++ {
				selectedPosition += itemHeights[i]
			}

			// If selected item is reasonably centered, use this window
			if selectedPosition >= currentLines/3 {
				break
			}
		}
	}

	return bestStart, bestEnd
}

// renderItemsRange renders items in the specified range with proper index mapping
func (m Model) renderItemsRange(items []string, start, end, selectedIndex int, itemHeights []int) string {
	var renderedItems []string

	for i := start; i < end && i < len(items); i++ {
		item := items[i]
		isSelected := (i == selectedIndex) // Используем глобальный индекс правильно

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
