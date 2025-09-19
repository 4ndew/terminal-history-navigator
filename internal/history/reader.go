package history

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Command represents a shell command with metadata
type Command struct {
	Text      string
	Position  int // Position in history file (higher = newer)
	Directory string
	Count     int
	ExitCode  int  // Exit code if available
	HasExit   bool // Whether exit code is available
}

// Reader handles reading command history from files
type Reader struct {
	sources         []string
	excludePatterns []*regexp.Regexp
	maxLines        int // Maximum lines to read from each file
}

// NewReader creates a new history reader with given sources
func NewReader(sources []string) *Reader {
	return &Reader{
		sources:  sources,
		maxLines: 5000, // Default limit
	}
}

// SetMaxLines sets the maximum number of lines to read from each file
func (r *Reader) SetMaxLines(maxLines int) {
	r.maxLines = maxLines
}

// SetExcludePatterns sets regex patterns for commands to exclude
func (r *Reader) SetExcludePatterns(patterns []string) error {
	r.excludePatterns = make([]*regexp.Regexp, 0, len(patterns))

	for _, pattern := range patterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		r.excludePatterns = append(r.excludePatterns, regex)
	}

	return nil
}

// ReadHistory reads command history from all configured sources
func (r *Reader) ReadHistory() ([]Command, error) {
	var allCommands []Command

	for _, source := range r.sources {
		// Check if file exists
		if _, err := os.Stat(source); os.IsNotExist(err) {
			continue
		}

		commands, err := r.readFromFile(source)
		if err != nil {
			continue // Skip problematic files but don't fail completely
		}

		allCommands = append(allCommands, commands...)
	}

	// Filter out problematic commands before sorting
	allCommands = r.filterProblematicCommands(allCommands)

	// Sort all commands by position (newest first - higher position = newer)
	sort.Slice(allCommands, func(i, j int) bool {
		return allCommands[i].Position > allCommands[j].Position
	})

	// Deduplicate and count frequency
	commandMap := make(map[string]*Command)
	var result []Command

	for _, cmd := range allCommands {
		// Skip excluded commands
		if r.shouldExclude(cmd.Text) {
			continue
		}

		// Clean command text
		cleanText := strings.TrimSpace(cmd.Text)
		if cleanText == "" {
			continue
		}

		if existing, found := commandMap[cleanText]; found {
			// Increment count and keep highest position (most recent appearance)
			existing.Count++
			if cmd.Position > existing.Position {
				existing.Position = cmd.Position
				existing.ExitCode = cmd.ExitCode
				existing.HasExit = cmd.HasExit
			}
		} else {
			// First occurrence - add to result and map
			newCmd := cmd
			newCmd.Text = cleanText
			newCmd.Count = 1
			commandMap[cleanText] = &newCmd
			result = append(result, newCmd)
		}
	}

	// Re-sort result by position after deduplication (newest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Position > result[j].Position
	})

	return result, nil
}

// filterProblematicCommands removes commands that cause display issues
func (r *Reader) filterProblematicCommands(commands []Command) []Command {
	var filtered []Command

	for _, cmd := range commands {
		// Skip commands that are clearly problematic
		if r.isProblematicCommand(cmd) {
			continue
		}
		filtered = append(filtered, cmd)
	}

	return filtered
}

// isProblematicCommand checks if a command should be filtered out
func (r *Reader) isProblematicCommand(cmd Command) bool {
	// Filter out empty commands
	cleanText := strings.TrimSpace(cmd.Text)
	if len(cleanText) == 0 {
		return true
	}

	// Filter out commands with control characters
	if strings.Contains(cleanText, "\x00") || strings.Contains(cleanText, "\xff") {
		return true
	}

	// Filter out commands that are just numbers
	if isJustNumber(cleanText) {
		return true
	}

	return false
}

// isJustNumber checks if a string contains only digits
func isJustNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// readFromFile reads commands from a specific history file
func (r *Reader) readFromFile(filename string) ([]Command, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read all lines
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Take only the last N lines (most recent commands)
	maxLines := r.maxLines
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	// Parse lines based on file type
	var commands []Command

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var cmd Command
		switch {
		case strings.Contains(filename, "zsh"):
			cmd = r.parseZshLine(line, i)
		case strings.Contains(filename, "bash") || filepath.Ext(filename) == ".bash_history":
			cmd = Command{
				Text:     strings.TrimSpace(line),
				Position: i, // Position in file
			}
		default:
			// Try zsh format first, then fallback
			if strings.HasPrefix(strings.TrimSpace(line), ":") {
				cmd = r.parseZshLine(line, i)
			} else {
				cmd = Command{
					Text:     strings.TrimSpace(line),
					Position: i, // Position in file
				}
			}
		}

		if cmd.Text != "" {
			commands = append(commands, cmd)
		}
	}

	return commands, nil
}

// parseZshLine parses a single zsh history line
func (r *Reader) parseZshLine(line string, lineNum int) Command {
	line = strings.TrimSpace(line)

	// Handle plain commands (not in zsh extended format)
	if !strings.HasPrefix(line, ":") {
		if line != "" {
			return Command{
				Text:     line,
				Position: lineNum, // Position in file
			}
		}
		return Command{Text: "", Position: lineNum}
	}

	// Extended zsh format: : timestamp:duration;command
	semiIndex := strings.Index(line, ";")
	if semiIndex == -1 || semiIndex == len(line)-1 {
		// Malformed line, try to extract command anyway
		if len(line) > 1 {
			possibleCmd := strings.TrimSpace(line[1:])
			if possibleCmd != "" && !strings.Contains(possibleCmd, ":") {
				return Command{
					Text:     possibleCmd,
					Position: lineNum,
				}
			}
		}
		return Command{Text: "", Position: lineNum}
	}

	// Extract metadata
	metadataPart := line[1:semiIndex]
	var exitCode int
	var hasExit bool

	parts := strings.Split(metadataPart, ":")
	// Check for exit code (third part in format timestamp:duration:exitcode)
	if len(parts) >= 3 && parts[2] != "" {
		if code, err := strconv.Atoi(parts[2]); err == nil {
			exitCode = code
			hasExit = true
		}
	}

	// Extract command
	command := strings.TrimSpace(line[semiIndex+1:])

	return Command{
		Text:     command,
		Position: lineNum, // Position in file
		ExitCode: exitCode,
		HasExit:  hasExit,
	}
}

// shouldExclude checks if a command should be excluded based on patterns
func (r *Reader) shouldExclude(command string) bool {
	for _, pattern := range r.excludePatterns {
		if pattern.MatchString(command) {
			return true
		}
	}
	return false
}
