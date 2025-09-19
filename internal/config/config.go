package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Sources         []string    `yaml:"sources"`
	ExcludePatterns []string    `yaml:"exclude_patterns"`
	UI              UIConfig    `yaml:"ui"`
	TemplatesPath   string      `yaml:"templates_path"`
	Performance     Performance `yaml:"performance"`
}

// UIConfig represents UI-specific settings
type UIConfig struct {
	MaxItems       int    `yaml:"max_items"`
	Theme          string `yaml:"theme"`
	ShowTimestamps bool   `yaml:"show_timestamps"`
	ShowFrequency  bool   `yaml:"show_frequency"`
}

// Performance represents performance-related settings
type Performance struct {
	CacheEnabled    bool `yaml:"cache_enabled"`
	MaxHistoryLines int  `yaml:"max_history_lines"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Sources: []string{
			filepath.Join(homeDir, ".zsh_history"),
			filepath.Join(homeDir, ".bash_history"),
		},
		ExcludePatterns: []string{
			"^sudo su", // Only sudo su commands (not all sudo)
			"password",
			"token",
			"secret",
			"key.*=", // Only environment variables with keys
			"^history",
			"^exit$",
			"^clear$",
			"^pwd$",
			"^\\.$",
			"^\\.\\.*$",
			"^\\d+$",         // Just numbers
			"^[[:space:]]*$", // Just whitespace
			"^h$",
		},
		UI: UIConfig{
			MaxItems:       1000,
			Theme:          "dark",
			ShowTimestamps: true,
			ShowFrequency:  true,
		},
		TemplatesPath: filepath.Join(homeDir, ".config", "history-nav", "templates.yaml"),
		Performance: Performance{
			CacheEnabled:    true,
			MaxHistoryLines: 10000,
		},
	}
}

// Load loads configuration from the config file or creates default config
func Load() (*Config, error) {
	configPath := getConfigPath()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		config := DefaultConfig()
		err := config.Save()
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := DefaultConfig()
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	// Expand home directory in paths
	config.expandPaths()

	return config, nil
}

// Save saves the configuration to the config file
func (c *Config) Save() error {
	configPath := getConfigPath()

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// expandPaths expands ~ to home directory in file paths
func (c *Config) expandPaths() {
	homeDir, _ := os.UserHomeDir()

	// Expand sources
	for i, source := range c.Sources {
		if strings.HasPrefix(source, "~/") {
			c.Sources[i] = filepath.Join(homeDir, source[2:])
		}
	}

	// Expand templates path
	if strings.HasPrefix(c.TemplatesPath, "~/") {
		c.TemplatesPath = filepath.Join(homeDir, c.TemplatesPath[2:])
	}
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "history-nav", "config.yaml")
}
