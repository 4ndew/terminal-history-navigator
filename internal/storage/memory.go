package storage

import (
	"sort"
	"strings"

	"github.com/4ndew/terminal-history-navigator/internal/history"
)

// Storage interface defines methods for storing and retrieving commands
type Storage interface {
	Store(commands []history.Command)
	Search(query string) []history.Command
	GetByFrequency() []history.Command
	GetRecent(limit int) []history.Command
	GetAll() []history.Command
}

// MemoryStorage implements in-memory storage for commands
type MemoryStorage struct {
	commands []history.Command
	indexed  map[string][]int // Maps words to command indices for fast search
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		commands: make([]history.Command, 0),
		indexed:  make(map[string][]int),
	}
}

// Store saves commands to memory and builds search index
func (s *MemoryStorage) Store(commands []history.Command) {
	s.commands = commands
	s.buildIndex()
}

// Search finds commands matching the query string
func (s *MemoryStorage) Search(query string) []history.Command {
	if query == "" {
		return s.GetRecent(1000) // Return recent commands if no query
	}

	query = strings.ToLower(query)
	matchingIndices := make(map[int]bool)

	// Split query into words
	queryWords := strings.Fields(query)

	for _, word := range queryWords {
		if indices, found := s.indexed[word]; found {
			for _, idx := range indices {
				matchingIndices[idx] = true
			}
		} else {
			// Fallback to substring search if word not in index
			for i, cmd := range s.commands {
				if strings.Contains(strings.ToLower(cmd.Text), word) {
					matchingIndices[i] = true
				}
			}
		}
	}

	// Collect matching commands
	results := make([]history.Command, 0, len(matchingIndices))
	for idx := range matchingIndices {
		results = append(results, s.commands[idx])
	}

	// Sort by relevance (frequency and recency)
	sort.Slice(results, func(i, j int) bool {
		// Primary sort by frequency
		if results[i].Count != results[j].Count {
			return results[i].Count > results[j].Count
		}
		// Secondary sort by timestamp
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	return results
}

// GetByFrequency returns commands sorted by usage frequency
func (s *MemoryStorage) GetByFrequency() []history.Command {
	commands := make([]history.Command, len(s.commands))
	copy(commands, s.commands)

	// Sort by frequency (count) first, then by recency
	sort.Slice(commands, func(i, j int) bool {
		// Primary sort by count (frequency) - higher count first
		if commands[i].Count != commands[j].Count {
			return commands[i].Count > commands[j].Count
		}
		// Secondary sort by timestamp - more recent first
		return commands[i].Timestamp.After(commands[j].Timestamp)
	})

	return commands
}

// GetRecent returns the most recently used commands
func (s *MemoryStorage) GetRecent(limit int) []history.Command {
	commands := make([]history.Command, len(s.commands))
	copy(commands, s.commands)

	// Sort by timestamp (most recent first)
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Timestamp.After(commands[j].Timestamp)
	})

	if limit > 0 && limit < len(commands) {
		commands = commands[:limit]
	}

	return commands
}

// GetAll returns all stored commands (sorted by recency)
func (s *MemoryStorage) GetAll() []history.Command {
	commands := make([]history.Command, len(s.commands))
	copy(commands, s.commands)

	// Sort by timestamp (most recent first) for consistency
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Timestamp.After(commands[j].Timestamp)
	})

	return commands
}

// buildIndex creates a search index for fast text searching
func (s *MemoryStorage) buildIndex() {
	s.indexed = make(map[string][]int)

	for i, cmd := range s.commands {
		// Index individual words from the command
		words := strings.Fields(strings.ToLower(cmd.Text))

		for _, word := range words {
			// Clean word of common shell characters
			word = cleanWord(word)
			if word == "" {
				continue
			}

			if _, exists := s.indexed[word]; !exists {
				s.indexed[word] = make([]int, 0)
			}
			s.indexed[word] = append(s.indexed[word], i)
		}

		// Also index command prefixes for partial matching
		cmdLower := strings.ToLower(cmd.Text)
		for j := 1; j <= len(cmdLower) && j <= 10; j++ {
			prefix := cmdLower[:j]
			if _, exists := s.indexed[prefix]; !exists {
				s.indexed[prefix] = make([]int, 0)
			}
			s.indexed[prefix] = append(s.indexed[prefix], i)
		}
	}
}

// cleanWord removes common shell characters from words
func cleanWord(word string) string {
	// Remove common shell characters
	word = strings.Trim(word, "\"'`()[]{}|&;")
	word = strings.TrimPrefix(word, "./")
	word = strings.TrimPrefix(word, "../")

	// Skip very short words and common shell operators
	if len(word) < 2 {
		return ""
	}

	// Skip common shell operators and flags
	switch word {
	case "&&", "||", ">>", "<<", "2>", "1>", "&>":
		return ""
	}

	return word
}

// GetStats returns storage statistics
func (s *MemoryStorage) GetStats() map[string]int {
	return map[string]int{
		"total_commands": len(s.commands),
		"unique_words":   len(s.indexed),
	}
}
