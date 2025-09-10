package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Theme           string           `json:"theme"`           // "dark" or "light"
	WindowWidth     float32          `json:"window_width"`
	WindowHeight    float32          `json:"window_height"`
	Argon2Defaults  Argon2Config     `json:"argon2_defaults"`
	LastUsedProfile string           `json:"last_used_profile"`
	History         []HistoryEntry   `json:"history"`
	Profiles        []Profile        `json:"profiles"`
}

// Argon2Config holds Argon2id parameters
type Argon2Config struct {
	Memory      uint32 `json:"memory"`      // in KiB
	Iterations  uint32 `json:"iterations"`
	Parallelism uint8  `json:"parallelism"`
}

// HistoryEntry represents a single operation in history
type HistoryEntry struct {
	FileName  string `json:"file_name"`
	Operation string `json:"operation"` // "encrypt" or "decrypt"
	Size      int64  `json:"size"`
	Timestamp int64  `json:"timestamp"` // Unix timestamp
	Result    string `json:"result"`    // "success" or "error"
	Error     string `json:"error,omitempty"`
}

// Profile represents a saved configuration preset
type Profile struct {
	Name            string `json:"name"`
	UseKeyfiles     bool   `json:"use_keyfiles"`
	ParanoidMode    bool   `json:"paranoid_mode"`
	ReedSolomon     bool   `json:"reed_solomon"`
	ForceDecrypt    bool   `json:"force_decrypt"`
	SplitOutput     bool   `json:"split_output"`
	CompressFiles   bool   `json:"compress_files"`
	DeniabilityMode bool   `json:"deniability_mode"`
	RecursiveMode   bool   `json:"recursive_mode"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Theme:        "dark",
		WindowWidth:  800,
		WindowHeight: 600,
		Argon2Defaults: Argon2Config{
			Memory:      64 * 1024, // 64 MiB
			Iterations:  1,
			Parallelism: 4,
		},
		LastUsedProfile: "",
		History:         []HistoryEntry{},
		Profiles: []Profile{
			{
				Name:            "Fast Archive",
				UseKeyfiles:     false,
				ParanoidMode:    false,
				ReedSolomon:     false,
				ForceDecrypt:    false,
				SplitOutput:     false,
				CompressFiles:   true,
				DeniabilityMode: false,
				RecursiveMode:   true,
			},
			{
				Name:            "Ultra Secure",
				UseKeyfiles:     true,
				ParanoidMode:    true,
				ReedSolomon:     true,
				ForceDecrypt:    false,
				SplitOutput:     false,
				CompressFiles:   false,
				DeniabilityMode: true,
				RecursiveMode:   false,
			},
			{
				Name:            "Cloud Upload",
				UseKeyfiles:     false,
				ParanoidMode:    false,
				ReedSolomon:     true,
				ForceDecrypt:    false,
				SplitOutput:     true,
				CompressFiles:   true,
				DeniabilityMode: false,
				RecursiveMode:   false,
			},
		},
	}
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".hadescrypt"), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), err
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return DefaultConfig(), err
	}

	return config, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// AddHistoryEntry adds a new entry to the history
func (c *Config) AddHistoryEntry(entry HistoryEntry) {
	c.History = append(c.History, entry)
	
	// Keep only the last 100 entries
	if len(c.History) > 100 {
		c.History = c.History[len(c.History)-100:]
	}
}

// ClearHistory removes all history entries
func (c *Config) ClearHistory() {
	c.History = []HistoryEntry{}
}

// GetProfile returns a profile by name, or nil if not found
func (c *Config) GetProfile(name string) *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			return &c.Profiles[i]
		}
	}
	return nil
}

// AddProfile adds or updates a profile
func (c *Config) AddProfile(profile Profile) {
	for i := range c.Profiles {
		if c.Profiles[i].Name == profile.Name {
			c.Profiles[i] = profile
			return
		}
	}
	c.Profiles = append(c.Profiles, profile)
}

// DeleteProfile removes a profile by name
func (c *Config) DeleteProfile(name string) {
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)
			return
		}
	}
}
