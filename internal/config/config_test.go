package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError error
	}{
		{
			name: "Valid config",
			config: Config{
				EngineVersion: "1.0.0",
				ExportVersion: "1.0.0",
				Salt:          "test_salt",
				Description:   "Test Export",
				HashType:      "sha256",
				Iterations:    1,
				Memory:        16,
			},
			wantError: nil,
		},
		{
			name: "Empty salt",
			config: Config{
				EngineVersion: "1.0.0",
				ExportVersion: "1.0.0",
				Description:   "Test Export",
				HashType:      "sha256",
			},
			wantError: ErrSaltEmpty,
		},
		{
			name: "Empty export version",
			config: Config{
				Salt:          "test_salt",
				EngineVersion: "1.0.0",
				Description:   "Test Export",
				HashType:      "sha256",
			},
			wantError: ErrExportVersionEmpty,
		},
		{
			name: "Empty engine version",
			config: Config{
				ExportVersion: "1.0.0",
				Salt:          "test_salt",
				Description:   "Test Export",
				HashType:      "sha256",
			},
			wantError: ErrEngineVersionEmpty,
		},
		{
			name: "Invalid hash type",
			config: Config{
				EngineVersion: "1.0.0",
				ExportVersion: "1.0.0",
				Salt:          "test_salt",
				Description:   "Test Export",
				HashType:      "invalid",
			},
			wantError: ErrInvalidHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadOrCreate(t *testing.T) {
	// Create temporary directory for tests
	tempDir := t.TempDir()

	t.Run("Config not found", func(t *testing.T) {
		config, err := LoadOrCreate(tempDir)
		assert.ErrorIs(t, err, ErrConfigNotFound)
		assert.Nil(t, config)
	})

	t.Run("Load existing config", func(t *testing.T) {
		// Create config with custom values
		existingConfig := &Config{
			EngineVersion: "1.0.0",
			ExportVersion: "2.0.0",
			Salt:          "custom_salt",
			Description:   "Custom Export",
			HashType:      "argon2id",
			Iterations:    2,
			Memory:        32,
		}
		err := existingConfig.Save(tempDir)
		require.NoError(t, err)

		// Load the config
		loadedConfig, err := LoadOrCreate(tempDir)
		require.NoError(t, err)

		// Verify loaded values match
		assert.Equal(t, existingConfig.EngineVersion, loadedConfig.EngineVersion)
		assert.Equal(t, existingConfig.ExportVersion, loadedConfig.ExportVersion)
		assert.Equal(t, existingConfig.Salt, loadedConfig.Salt)
		assert.Equal(t, existingConfig.Description, loadedConfig.Description)
		assert.Equal(t, existingConfig.HashType, loadedConfig.HashType)
		assert.Equal(t, existingConfig.Iterations, loadedConfig.Iterations)
		assert.Equal(t, existingConfig.Memory, loadedConfig.Memory)
	})

	t.Run("Invalid directory", func(t *testing.T) {
		_, err := LoadOrCreate("/nonexistent/directory")
		assert.Error(t, err)
	})
}

func TestConfig_Save(t *testing.T) {
	tempDir := t.TempDir()

	config := &Config{
		EngineVersion: "1.0.0",
		ExportVersion: "1.0.0",
		Salt:          "test_salt",
		Description:   "Test Export",
		HashType:      "sha256",
		Iterations:    1,
		Memory:        16,
	}

	t.Run("Save valid config", func(t *testing.T) {
		err := config.Save(tempDir)
		assert.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(filepath.Join(tempDir, configFileName))
		assert.NoError(t, err)

		// Load and verify contents
		loadedConfig, err := Load(tempDir)
		require.NoError(t, err)
		assert.Equal(t, config, loadedConfig)
	})

	t.Run("Save to invalid directory", func(t *testing.T) {
		err := config.Save("/nonexistent/directory")
		assert.Error(t, err)
	})

	t.Run("Save invalid config", func(t *testing.T) {
		invalidConfig := &Config{} // Empty config
		err := invalidConfig.Save(tempDir)
		assert.Error(t, err)
	})
}
