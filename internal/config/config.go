package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	RefreshInterval time.Duration `yaml:"refresh_interval"`
	LogTailLines    int           `yaml:"log_tail_lines"`
	Theme           string        `yaml:"theme"`
	Favorites       Favorites     `yaml:"favorites"`
	KeyBindings     KeyBindings   `yaml:"key_bindings"`
	Display         Display       `yaml:"display"`
}

// Favorites contains user's favorite namespaces and contexts
type Favorites struct {
	Namespaces []string `yaml:"namespaces"`
	Contexts   []string `yaml:"contexts"`
}

// KeyBindings defines customizable keyboard shortcuts
type KeyBindings struct {
	Logs        string `yaml:"logs"`
	Delete      string `yaml:"delete"`
	Describe    string `yaml:"describe"`
	Exec        string `yaml:"exec"`
	PortForward string `yaml:"port_forward"`
	Refresh     string `yaml:"refresh"`
	Filter      string `yaml:"filter"`
	Search      string `yaml:"search"`
	Help        string `yaml:"help"`
}

// Display contains display preferences
type Display struct {
	ShowTimestamps bool `yaml:"show_timestamps"`
	CompactMode    bool `yaml:"compact_mode"`
	MaxNameLength  int  `yaml:"max_name_length"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		RefreshInterval: 5 * time.Second,
		LogTailLines:    100,
		Theme:           "dark",
		Favorites: Favorites{
			Namespaces: []string{},
			Contexts:   []string{},
		},
		KeyBindings: KeyBindings{
			Logs:        "l",
			Delete:      "d",
			Describe:    "D",
			Exec:        "e",
			PortForward: "p",
			Refresh:     "r",
			Filter:      "f",
			Search:      "/",
			Help:        "?",
		},
		Display: Display{
			ShowTimestamps: true,
			CompactMode:    false,
			MaxNameLength:  50,
		},
	}
}

// Load reads the configuration from the config file
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return DefaultConfig(), nil
		}
		configPath = filepath.Join(home, ".kubetui", "config.yaml")
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the configuration to the config file
func (c *Config) Save(configPath string) error {
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configPath = filepath.Join(home, ".kubetui", "config.yaml")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
