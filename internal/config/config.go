package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configFileName = "export_config.json"
)

var (
	ErrConfigNotFound     = errors.New("config not found")
	ErrSaltEmpty          = errors.New("salt cannot be empty")
	ErrExportVersionEmpty = errors.New("export version cannot be empty")
	ErrEngineVersionEmpty = errors.New("engine version cannot be empty")
	ErrInvalidHash        = errors.New("invalid hash type")
)

// Config represents the export configuration.
type Config struct {
	EngineVersion string `json:"engineVersion"` // Version of the engine
	ExportVersion string `json:"exportVersion"` // Version of the export
	Salt          string `json:"salt"`          // Salt used for hashing IDs
	Description   string `json:"description"`   // Description of the export
	HashType      string `json:"hashType"`      // Type of hash algorithm to use
	Iterations    uint32 `json:"iterations"`    // Number of iterations for hashing
	Memory        uint32 `json:"memory"`        // Memory parameter for Argon2id (in MB)
}

// LoadOrCreate loads the configuration from the specified directory.
func LoadOrCreate(dir string) (*Config, error) {
	// Try to load existing config
	config, err := Load(dir)
	if err == nil {
		// Validate existing config
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("invalid configuration: %w", err)
		}
		return config, nil
	}

	// If config doesn't exist, create default
	if os.IsNotExist(err) {
		return nil, ErrConfigNotFound
	}

	return nil, fmt.Errorf("failed to load config: %w", err)
}

// Load reads the configuration from the specified directory.
func Load(dir string) (*Config, error) {
	configPath := filepath.Join(dir, configFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// Save writes the configuration to the specified directory.
func (c *Config) Save(dir string) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	configPath := filepath.Join(dir, configFileName)

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.EngineVersion == "" {
		return ErrEngineVersionEmpty
	}
	if c.ExportVersion == "" {
		return ErrExportVersionEmpty
	}
	if c.Salt == "" {
		return ErrSaltEmpty
	}
	if c.HashType != "" && c.HashType != "argon2id" && c.HashType != "sha256" {
		return ErrInvalidHash
	}
	return nil
}
