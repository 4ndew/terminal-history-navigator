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

// Command represents a shell command with metadata.
type Command struct {
	Text      string
	Timestamp int64 // unix seconds; synthesized from file mtime when the source has no timestamps
	Count     int
}

// Reader handles reading command history from files.
type Reader struct {
	sources         []string
	excludePatterns []*regexp.Regexp
	maxLines        int
}

// NewReader creates a new history reader with given sources.
func NewReader(sources []string) *Reader {
	return &Reader{
		sources:  sources,
		maxLines: 5000,
	}
}

// SetMaxLines sets the maximum number of logical lines to keep from each file.
func (r *Reader) SetMaxLines(maxLines int) {
	r.maxLines = maxLines
}

// SetExcludePatterns sets regex patterns for commands to exclude.
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

// ReadHistory reads command history from all configured sources,
// deduplicates commands, counts real frequency and sorts newest first.
func (r *Reader) ReadHistory() ([]Command, error) {
	var all []Command

	for _, source := range r.sources {
		if _, err := os.Stat(source); os.IsNotExist(err) {
			continue
		}
		commands, err := r.readFromFile(source)
		if err != nil {
			continue // skip problematic files but don't fail completely
		}
		all = append(all, commands...)
	}

	// Deduplicate: count real frequency, keep the newest timestamp per command.
	index := make(map[string]int)
	var result []Command

	for _, cmd := range all {
		text := strings.TrimSpace(cmd.Text)
		if text == "" || isProblematic(text) || r.shouldExclude(text) {
			continue
		}

		if i, ok := index[text]; ok {
			result[i].Count++
			if cmd.Timestamp > result[i].Timestamp {
				result[i].Timestamp = cmd.Timestamp
			}
		} else {
			index[text] = len(result)
			result = append(result, Command{
				Text:      text,
				Timestamp: cmd.Timestamp,
				Count:     1,
			})
		}
	}

	// Newest first.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp > result[j].Timestamp
	})

	return result, nil
}

// readFromFile reads commands from a specific history file.
func (r *Reader) readFromFile(filename string) ([]Command, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Default scanner limit is 64KB per line; one long line would
	// otherwise abort reading the whole file.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var physical []string
	for scanner.Scan() {
		physical = append(physical, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Join backslash-continued lines into logical multiline commands.
	logical := joinContinuations(physical)

	// Keep only the last N logical lines (most recent commands).
	if len(logical) > r.maxLines {
		logical = logical[len(logical)-r.maxLines:]
	}

	isZsh := strings.Contains(filepath.Base(filename), "zsh")

	var commands []Command
	var pendingTS int64

	for _, line := range logical {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// bash with HISTTIMEFORMAT writes "#<unix-ts>" before each command.
		if ts, ok := parseBashTimestamp(trimmed); ok {
			pendingTS = ts
			continue
		}

		var cmd Command
		if isZsh || strings.HasPrefix(trimmed, ": ") {
			cmd = parseZshLine(trimmed)
		} else {
			cmd = Command{Text: trimmed, Timestamp: pendingTS}
		}
		pendingTS = 0

		if strings.TrimSpace(cmd.Text) != "" {
			commands = append(commands, cmd)
		}
	}

	// Files (or entries) without timestamps get synthesized ones based on
	// file mtime, so they can be merged and ordered with timestamped history.
	fillMissingTimestamps(commands, fileMTime(filename))

	return commands, nil
}

// joinContinuations merges physical lines ending with a backslash into one
// logical line, preserving real newlines inside the command.
// Limitation: a command whose stored form legitimately ends with a literal
// backslash will be merged with the next line (heuristic).
func joinContinuations(lines []string) []string {
	var out []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		for strings.HasSuffix(line, "\\") && i+1 < len(lines) {
			line = strings.TrimSuffix(line, "\\") + "\n" + lines[i+1]
			i++
		}
		out = append(out, line)
	}
	return out
}

var bashTSPattern = regexp.MustCompile(`^#(\d+)$`)

// parseBashTimestamp recognizes bash HISTTIMEFORMAT comment lines like "#1700000000".
func parseBashTimestamp(line string) (int64, bool) {
	m := bashTSPattern.FindStringSubmatch(line)
	if m == nil {
		return 0, false
	}
	ts, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil || ts < 100000000 { // sanity check: must look like a unix timestamp
		return 0, false
	}
	return ts, true
}

// parseZshLine parses a zsh history entry.
// Extended format: ": <timestamp>:<elapsed>;command". The second field is the
// command duration, NOT an exit code — zsh does not store exit codes.
func parseZshLine(line string) Command {
	if !strings.HasPrefix(line, ":") {
		return Command{Text: line}
	}

	semi := strings.Index(line, ";")
	if semi == -1 || semi == len(line)-1 {
		return Command{}
	}

	meta := strings.TrimPrefix(line[:semi], ":")
	var ts int64
	parts := strings.Split(meta, ":")
	if len(parts) >= 1 {
		if v, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64); err == nil && v > 0 {
			ts = v
		}
	}

	return Command{
		Text:      strings.TrimSpace(line[semi+1:]),
		Timestamp: ts,
	}
}

// fillMissingTimestamps assigns approximate timestamps to entries that have
// none: the last entry gets the file mtime, earlier entries step back 1s each.
func fillMissingTimestamps(commands []Command, mtime int64) {
	n := len(commands)
	for i := range commands {
		if commands[i].Timestamp == 0 {
			commands[i].Timestamp = mtime - int64(n-1-i)
		}
	}
}

// fileMTime returns the file modification time in unix seconds,
// falling back to now if stat fails.
func fileMTime(filename string) int64 {
	if fi, err := os.Stat(filename); err == nil {
		return fi.ModTime().Unix()
	}
	return time.Now().Unix()
}

// isProblematic checks if a command should be filtered out.
func isProblematic(text string) bool {
	if strings.Contains(text, "\x00") || strings.Contains(text, "\xff") {
		return true
	}
	if isJustNumber(text) {
		return true
	}
	return false
}

// isJustNumber checks if a string contains only digits.
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

// shouldExclude checks if a command should be excluded based on patterns.
func (r *Reader) shouldExclude(command string) bool {
	for _, pattern := range r.excludePatterns {
		if pattern.MatchString(command) {
			return true
		}
	}
	return false
}