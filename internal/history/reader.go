package history

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Command represents a shell command with metadata
type Command struct {
	Text      string
	Timestamp time.Time
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

	// Sort all commands by timestamp (newest first)
	sort.Slice(allCommands, func(i, j int) bool {
		return allCommands[i].Timestamp.After(allCommands[j].Timestamp)
	})

	// Deduplicate and count, keeping the most recent occurrence at the top
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
			// Update count and keep most recent timestamp
			existing.Count++
			if cmd.Timestamp.After(existing.Timestamp) {
				existing.Timestamp = cmd.Timestamp
				existing.ExitCode = cmd.ExitCode
				existing.HasExit = cmd.HasExit
			}
		} else {
			// First occurrence - add to result and map
			cmd.Text = cleanText
			cmd.Count = 1
			commandMap[cleanText] = &cmd
			result = append(result, cmd)
		}
	}

	// Re-sort result by timestamp after deduplication to ensure proper order
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})

	return result, nil
}

// readFromFile reads commands from a specific history file
func (r *Reader) readFromFile(filename string) ([]Command, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read all lines first
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
	ext := filepath.Ext(filename)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var cmd Command
		switch {
		case strings.Contains(filename, "zsh"):
			cmd = r.parseZshLine(line)
		case strings.Contains(filename, "bash") || ext == ".bash_history":
			cmd = Command{
				Text:      strings.TrimSpace(line),
				Timestamp: time.Now(),
			}
		default:
			// Try zsh format first, then fallback
			if strings.HasPrefix(line, ":") {
				cmd = r.parseZshLine(line)
			} else {
				cmd = Command{
					Text:      strings.TrimSpace(line),
					Timestamp: time.Now(),
				}
			}
		}

		if cmd.Text != "" {
			commands = append(commands, cmd)
		}
	}

	return commands, nil
}

// parseZshHistory parses zsh history format
func (r *Reader) parseZshHistory(scanner *bufio.Scanner) []Command {
	var commands []Command
	lineCount := 0

	for scanner.Scan() && lineCount < r.maxLines {
		line := scanner.Text()
		lineCount++

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		cmd := r.parseZshLine(line)
		if cmd.Text != "" {
			commands = append(commands, cmd)
		}
	}

	return commands
}

// parseBashHistory parses bash history format
func (r *Reader) parseBashHistory(scanner *bufio.Scanner) []Command {
	var commands []Command
	lineCount := 0

	for scanner.Scan() && lineCount < r.maxLines {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		// Skip empty lines
		if line == "" {
			continue
		}

		// Bash history is typically just command text
		cmd := Command{
			Text:      line,
			Timestamp: time.Now(), // No timestamp in basic bash history
		}

		commands = append(commands, cmd)
	}

	return commands
}

// parseGenericHistory tries to parse unknown history format
func (r *Reader) parseGenericHistory(scanner *bufio.Scanner) []Command {
	var commands []Command

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		// Try zsh format first
		if strings.HasPrefix(line, ":") {
			cmd := r.parseZshLine(line)
			if cmd.Text != "" {
				commands = append(commands, cmd)
				continue
			}
		}

		// Fall back to treating as plain command
		cmd := Command{
			Text:      line,
			Timestamp: time.Now(),
		}
		commands = append(commands, cmd)
	}

	return commands
}

// parseZshLine parses a single zsh history line
func (r *Reader) parseZshLine(line string) Command {
	line = strings.TrimSpace(line)

	// Handle multi-line commands (zsh can have continuation lines)
	if !strings.HasPrefix(line, ":") {
		return Command{
			Text:      line,
			Timestamp: time.Now(),
		}
	}

	// Extended zsh format can include exit code:
	// : 1640995200:0;command (standard)
	// : 1640995200:0:1;command (with exit code 1)

	// Find the semicolon that separates metadata from command
	semiIndex := strings.Index(line, ";")
	if semiIndex == -1 || semiIndex == len(line)-1 {
		// No command after semicolon
		return Command{Text: "", Timestamp: time.Now()}
	}

	// Extract metadata from the beginning
	metadataPart := line[1:semiIndex] // Remove leading ':'
	var timestamp time.Time
	var exitCode int
	var hasExit bool

	// Split by colon to get timestamp, duration, and potentially exit code
	parts := strings.Split(metadataPart, ":")
	if len(parts) >= 1 && parts[0] != "" {
		if ts, err := parseTimestamp(parts[0]); err == nil {
			timestamp = ts
		} else {
			timestamp = time.Now()
		}
	} else {
		timestamp = time.Now()
	}

	// Check for exit code (some zsh configurations store this)
	if len(parts) >= 3 && parts[2] != "" {
		if code, err := strconv.Atoi(parts[2]); err == nil {
			exitCode = code
			hasExit = true
		}
	}

	// Extract command (everything after semicolon)
	command := strings.TrimSpace(line[semiIndex+1:])

	return Command{
		Text:      command,
		Timestamp: timestamp,
		ExitCode:  exitCode,
		HasExit:   hasExit,
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

// parseTimestamp parses a Unix timestamp string
func parseTimestamp(timestampStr string) (time.Time, error) {
	// Try parsing as Unix timestamp
	if len(timestampStr) == 10 {
		// Parse Unix timestamp (seconds)
		if timestamp, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
			return time.Unix(timestamp, 0), nil
		}
	}

	// Fallback to current time
	return time.Now(), nil
}
