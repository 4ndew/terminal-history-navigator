package history

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
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
}

// Reader handles reading command history from files
type Reader struct {
	sources         []string
	excludePatterns []*regexp.Regexp
}

// NewReader creates a new history reader with given sources
func NewReader(sources []string) *Reader {
	return &Reader{
		sources: sources,
	}
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
	commandMap := make(map[string]*Command)

	for _, source := range r.sources {
		// Check if file exists
		if _, err := os.Stat(source); os.IsNotExist(err) {
			continue
		}

		commands, err := r.readFromFile(source)
		if err != nil {
			continue // Skip problematic files but don't fail completely
		}

		// Merge commands and count duplicates
		for _, cmd := range commands {
			if existing, found := commandMap[cmd.Text]; found {
				existing.Count++
				// Keep the most recent timestamp
				if cmd.Timestamp.After(existing.Timestamp) {
					existing.Timestamp = cmd.Timestamp
					existing.Directory = cmd.Directory
				}
			} else {
				cmd.Count = 1
				commandMap[cmd.Text] = &cmd
			}
		}
	}

	// Convert map to slice
	result := make([]Command, 0, len(commandMap))
	for _, cmd := range commandMap {
		if !r.shouldExclude(cmd.Text) {
			result = append(result, *cmd)
		}
	}

	return result, nil
}

// readFromFile reads commands from a specific history file
func (r *Reader) readFromFile(filename string) ([]Command, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []Command
	scanner := bufio.NewScanner(file)

	// Determine file type by extension
	ext := filepath.Ext(filename)
	switch {
	case strings.Contains(filename, "zsh"):
		commands = r.parseZshHistory(scanner)
	case strings.Contains(filename, "bash") || ext == ".bash_history":
		commands = r.parseBashHistory(scanner)
	default:
		// Try to auto-detect format
		commands = r.parseGenericHistory(scanner)
	}

	return commands, scanner.Err()
}

// parseZshHistory parses zsh history format
func (r *Reader) parseZshHistory(scanner *bufio.Scanner) []Command {
	var commands []Command

	for scanner.Scan() {
		line := scanner.Text()

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

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

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
	// Zsh format: ": timestamp:elapsed;command"
	if !strings.HasPrefix(line, ":") {
		return Command{Text: line, Timestamp: time.Now()}
	}

	// Find the first semicolon
	semiIndex := strings.Index(line, ";")
	if semiIndex == -1 {
		return Command{Text: line, Timestamp: time.Now()}
	}

	// Extract timestamp
	timestampPart := line[1:semiIndex]
	colonIndex := strings.Index(timestampPart, ":")

	var timestamp time.Time
	if colonIndex != -1 {
		timestampStr := timestampPart[:colonIndex]
		if ts, err := parseTimestamp(timestampStr); err == nil {
			timestamp = ts
		} else {
			timestamp = time.Now()
		}
	} else {
		timestamp = time.Now()
	}

	// Extract command
	command := line[semiIndex+1:]

	return Command{
		Text:      command,
		Timestamp: timestamp,
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
