package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	// Colors
	primaryColor = lipgloss.Color("#00D4AA")
	accentColor  = lipgloss.Color("#F59E0B")
	mutedColor   = lipgloss.Color("#6B7280")
	errorColor   = lipgloss.Color("#EF4444")

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Item styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#4A5568")).
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

// View renders the TUI interface.
func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	var sections []string

	sections = append(sections, m.renderHeader())
	sections = append(sections, "")
	sections = append(sections, m.renderMainContent())
	sections = append(sections, "")
	sections = append(sections, m.renderFooter())

	return strings.Join(sections, "\n")
}

// renderHeader renders the application header.
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

// renderMainContent renders the main content area with scrolling support
// for items that wrap onto multiple lines.
func (m Model) renderMainContent() string {
	items, selectedIndex := m.getVisibleItems()

	if len(items) == 0 {
		return m.renderEmptyState()
	}

	maxVisibleLines := m.height - 6 // Header(1) + separator(1) + separator(1) + footer(3)
	if maxVisibleLines < 3 {
		maxVisibleLines = 3
	}

	itemHeights := make([]int, len(items))
	totalLines := 0
	for i, item := range items {
		height := m.calculateItemHeight(item)
		itemHeights[i] = height
		totalLines += height
	}

	if totalLines <= maxVisibleLines {
		return m.renderItemsRange(items, 0, len(items), selectedIndex)
	}

	m.adjustScrollOffset(itemHeights, maxVisibleLines)

	lines := 0
	end := m.scrollOffset
	for i := m.scrollOffset; i < len(items); i++ {
		lines += itemHeights[i]
		if lines > maxVisibleLines {
			break
		}
		end = i + 1
	}

	return m.renderItemsRange(items, m.scrollOffset, end, selectedIndex)
}

// calculateItemHeight calculates how many lines an item will occupy.
// Counts runes; assumes one terminal cell per rune (wide CJK glyphs
// may be slightly underestimated).
func (m Model) calculateItemHeight(item string) int {
	maxWidth := m.width - 6
	if maxWidth < 20 {
		maxWidth = 20
	}

	availableForText := maxWidth - 2 // prefix "► " / "  "
	if availableForText < 10 {
		availableForText = 10
	}

	if utf8.RuneCountInString(item) <= availableForText {
		return 1
	}

	return len(wrapText(item, availableForText))
}

// renderItemsRange renders items in the specified range.
func (m Model) renderItemsRange(items []string, start, end, selectedIndex int) string {
	var renderedItems []string

	for i := start; i < end && i < len(items); i++ {
		renderedItems = append(renderedItems, m.renderSingleItem(items[i], i == selectedIndex))
	}

	return strings.Join(renderedItems, "\n")
}

// renderSingleItem renders a single item with proper wrapping.
func (m Model) renderSingleItem(item string, isSelected bool) string {
	maxWidth := m.width - 6
	if maxWidth < 20 {
		maxWidth = 20
	}

	var prefix string
	if isSelected {
		prefix = "► "
	} else {
		prefix = "  "
	}

	style := normalItemStyle
	if isSelected {
		style = selectedItemStyle
	}

	// Fits in one line.
	if utf8.RuneCountInString(prefix+item) <= maxWidth {
		return style.Render(prefix + item)
	}

	availableForText := maxWidth - 2
	if availableForText < 10 {
		availableForText = 10
	}

	lines := wrapText(item, availableForText)
	var wrappedLines []string
	for j, line := range lines {
		linePrefix := "  "
		if j == 0 {
			linePrefix = prefix
		}
		wrappedLines = append(wrappedLines, style.Render(linePrefix+line))
	}

	return strings.Join(wrappedLines, "\n")
}

// wrapText wraps text to the specified width in runes (UTF-8 safe).
func wrapText(text string, width int) []string {
	runes := []rune(text)
	if len(runes) <= width {
		return []string{text}
	}

	var lines []string
	for len(runes) > 0 {
		if len(runes) <= width {
			lines = append(lines, strings.TrimSpace(string(runes)))
			break
		}

		// Find best break point (prefer a space in the second half).
		breakPoint := width
		for i := width - 1; i >= width/2 && i > 0; i-- {
			if runes[i] == ' ' {
				breakPoint = i
				break
			}
		}

		lines = append(lines, strings.TrimSpace(string(runes[:breakPoint])))
		runes = []rune(strings.TrimSpace(string(runes[breakPoint:])))
	}

	return lines
}

// renderEmptyState renders the empty state message.
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

// renderFooter renders the footer with status and controls.
func (m Model) renderFooter() string {
	var sections []string

	if m.errorMsg != "" {
		sections = append(sections, errorStyle.Render("Error: "+m.errorMsg))
	} else if m.statusMsg != "" {
		sections = append(sections, statusStyle.Render(m.statusMsg))
	}

	itemCount := m.getItemCount()
	if itemCount > 0 {
		position := fmt.Sprintf("%d/%d", m.cursor+1, itemCount)

		var sortInfo string
		if m.mode == HistoryMode {
			if m.sortByFreq {
				sortInfo = " (by frequency)"
			} else {
				sortInfo = " (newest first)"
			}
		}

		sections = append(sections, lipgloss.NewStyle().Foreground(mutedColor).Render(position+sortInfo))
	}

	controls := m.getControlsHelp()
	sections = append(sections, footerStyle.Render(controls))

	footer := strings.Join(sections, " | ")
	return m.wrapFooter(footer)
}

// wrapFooter wraps the footer text if it exceeds screen width.
// Uses lipgloss.Width to measure visible width (ignores ANSI codes).
func (m Model) wrapFooter(footer string) string {
	maxWidth := m.width - 4
	if maxWidth < 20 {
		maxWidth = 20
	}

	if lipgloss.Width(footer) <= maxWidth {
		return footer
	}

	parts := strings.Split(footer, " | ")
	var lines []string
	var currentLine string

	for i, part := range parts {
		testLine := currentLine
		if testLine != "" {
			testLine += " | "
		}
		testLine += part

		if lipgloss.Width(testLine) <= maxWidth {
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

// getControlsHelp returns context-appropriate control hints.
func (m Model) getControlsHelp() string {
	switch m.mode {
	case SearchMode:
		return "esc: exit | enter: copy | ↑↓: navigate"
	case TemplatesMode:
		return "enter: copy | d: delete | t: history | /: search | ?: help | q: quit"
	default:
		return "enter: copy | s: save as template | t: templates | /: search | f: freq | ?: help | q: quit"
	}
}

// renderHelp renders the help screen.
func (m Model) renderHelp() string {
	helpText := `Terminal History Navigator - Help

NAVIGATION:
  ↑/k         Move up
  ↓/j         Move down
  mouse wheel Scroll list
  enter       Copy selected item to clipboard

MODES:
  h           Switch to history mode
  t           Toggle templates mode
  /           Start search
  f           Toggle frequency / chronological sort (history mode)

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