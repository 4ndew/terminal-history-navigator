package storage

import (
	"sort"
	"strings"

	"github.com/4ndew/terminal-history-navigator/internal/history"
)

// Storage interface defines methods for storing and retrieving commands.
type Storage interface {
	Store(commands []history.Command)
	Search(query string) []history.Command
	GetByFrequency() []history.Command
	GetRecent(limit int) []history.Command
	GetAll() []history.Command
}

// MemoryStorage implements in-memory storage for commands.
type MemoryStorage struct {
	commands []history.Command
}

// NewMemoryStorage creates a new in-memory storage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		commands: make([]history.Command, 0),
	}
}

// Store saves commands to memory.
func (s *MemoryStorage) Store(commands []history.Command) {
	s.commands = commands
}

// Search finds commands containing all query words as whole words or prefixes.
func (s *MemoryStorage) Search(query string) []history.Command {
	if query == "" {
		return s.GetRecent(1000)
	}

	queryWords := strings.Fields(strings.ToLower(query))

	var results []history.Command
	for _, cmd := range s.commands {
		if commandMatchesQuery(strings.ToLower(cmd.Text), queryWords) {
			results = append(results, cmd)
		}
	}

	// Newest first.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp > results[j].Timestamp
	})

	return results
}

// commandMatchesQuery checks if a command matches all query words.
func commandMatchesQuery(cmdText string, queryWords []string) bool {
	cmdWords := strings.Fields(cmdText)
	for _, queryWord := range queryWords {
		if !commandContainsWord(cmdWords, queryWord) {
			return false
		}
	}
	return true
}

// commandContainsWord checks if command contains a word as whole word or prefix.
func commandContainsWord(cmdWords []string, queryWord string) bool {
	for _, cmdWord := range cmdWords {
		cleanCmdWord := cleanWord(cmdWord)
		if cleanCmdWord != "" {
			if cleanCmdWord == queryWord || strings.HasPrefix(cleanCmdWord, queryWord) {
				return true
			}
		}
		if strings.HasPrefix(strings.ToLower(cmdWord), queryWord) {
			return true
		}
	}
	return false
}

// GetByFrequency returns all commands sorted by real usage count,
// then by recency for equal counts.
func (s *MemoryStorage) GetByFrequency() []history.Command {
	commands := make([]history.Command, len(s.commands))
	copy(commands, s.commands)

	sort.Slice(commands, func(i, j int) bool {
		if commands[i].Count != commands[j].Count {
			return commands[i].Count > commands[j].Count
		}
		return commands[i].Timestamp > commands[j].Timestamp
	})

	return commands
}

// GetRecent returns the most recently used commands (newest first).
func (s *MemoryStorage) GetRecent(limit int) []history.Command {
	commands := make([]history.Command, len(s.commands))
	copy(commands, s.commands)

	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Timestamp > commands[j].Timestamp
	})

	if limit > 0 && limit < len(commands) {
		commands = commands[:limit]
	}

	return commands
}

// GetAll returns all stored commands (newest first).
func (s *MemoryStorage) GetAll() []history.Command {
	return s.GetRecent(0)
}

// cleanWord removes common shell characters from words.
func cleanWord(word string) string {
	word = strings.Trim(word, "\"'`()[]{}|&;")
	word = strings.TrimPrefix(word, "./")
	word = strings.TrimPrefix(word, "../")

	if len(word) < 2 {
		return ""
	}

	switch word {
	case "&&", "||", ">>", "<<", "2>", "1>", "&>":
		return ""
	}

	return word
}