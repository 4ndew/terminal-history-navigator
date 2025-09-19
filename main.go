package main

import (
	"fmt"
	"log"
	"os"

	"github.com/4ndew/terminal-history-navigator/internal/config"
	"github.com/4ndew/terminal-history-navigator/internal/history"
	"github.com/4ndew/terminal-history-navigator/internal/storage"
	"github.com/4ndew/terminal-history-navigator/internal/templates"
	"github.com/4ndew/terminal-history-navigator/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize storage
	store := storage.NewMemoryStorage()

	// Initialize reader
	reader := history.NewReader(cfg.Sources)
	reader.SetMaxLines(cfg.Performance.MaxHistoryLines)

	// Set exclude patterns if any configured
	if len(cfg.ExcludePatterns) > 0 {
		err = reader.SetExcludePatterns(cfg.ExcludePatterns)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Invalid exclude patterns: %v\n", err)
		}
	}

	// Load initial history
	err = loadHistory(reader, store)
	if err != nil {
		log.Fatalf("Failed to read history: %v", err)
	}

	// Load templates
	templateLoader := templates.NewLoader(cfg.TemplatesPath)
	templatesData, err := templateLoader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load templates: %v\n", err)
		// Continue without templates
	}

	// Create refresh callback
	refreshData := func() error {
		return loadHistory(reader, store)
	}

	// Create UI model with refresh callback
	model := ui.NewModel(store, templatesData, cfg, refreshData)

	// Create TUI program
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

// loadHistory reads command history and stores it
func loadHistory(reader *history.Reader, store storage.Storage) error {
	commands, err := reader.ReadHistory()
	if err != nil {
		return err
	}

	// Store commands
	store.Store(commands)
	return nil
}
