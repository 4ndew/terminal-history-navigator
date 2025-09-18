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

	// Read command history
	reader := history.NewReader(cfg.Sources)
	reader.SetMaxLines(cfg.Performance.MaxHistoryLines)

	// Set exclude patterns if any configured
	if len(cfg.ExcludePatterns) > 0 {
		err = reader.SetExcludePatterns(cfg.ExcludePatterns)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Invalid exclude patterns: %v\n", err)
		}
	}

	commands, err := reader.ReadHistory()
	if err != nil {
		log.Fatalf("Failed to read history: %v", err)
	}

	// Store commands
	store.Store(commands)

	// Load templates
	templateLoader := templates.NewLoader(cfg.TemplatesPath)
	templatesData, err := templateLoader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load templates: %v\n", err)
		// Continue without templates
	}

	// Create UI model
	model := ui.NewModel(store, templatesData, cfg)

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
