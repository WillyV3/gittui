package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	goghthemes "github.com/willyv3/gogh-themes"
)

// Config represents the application configuration
type Config struct {
	Username               string   `json:"username"`                   // GitHub username
	Theme                  string   `json:"theme"`                      // Theme name from gogh-themes (361 available)
	AvatarWidth            int      `json:"avatar_width"`               // Avatar width in characters (30-240, default 60)
	AvatarSource           string   `json:"avatar_source"`              // Avatar source: "personal" or org name
	AvatarRenderer         string   `json:"avatar_renderer"`            // Avatar renderer: "mosaic" or "ascii"
	AllowedOrgs            []string `json:"allowed_orgs"`               // Organizations to include private repos from
	IncludePersonalPrivate bool     `json:"include_personal_private"`   // Include your personal private repos
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Username:               "",          // No default username
		Theme:                  "Dracula",   // Default theme
		AvatarWidth:            60,          // Default avatar width
		AvatarSource:           "personal",  // Default to personal avatar
		AvatarRenderer:         "mosaic",    // Default to mosaic renderer
		AllowedOrgs:            []string{},  // Empty = no org repos included
		IncludePersonalPrivate: false,       // Default to public only
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".local", "gittui")
	configPath := filepath.Join(configDir, "config.json")

	return configPath, nil
}

// LoadConfig loads the configuration from ~/.local/gittui/config.json
// If the config doesn't exist, it creates one with defaults
func LoadConfig() (Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	// Check if config exists
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config doesn't exist - create it with defaults
			config := DefaultConfig()
			if saveErr := SaveConfig(config); saveErr != nil {
				// Non-fatal: return default config even if save fails
				fmt.Fprintf(os.Stderr, "Warning: Failed to create config file: %v\n", saveErr)
			}
			return config, nil
		}
		// Some other read error
		return DefaultConfig(), fmt.Errorf("failed to read config: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig(), fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate theme exists
	if _, ok := goghthemes.Get(config.Theme); !ok {
		fmt.Fprintf(os.Stderr, "Warning: Theme '%s' not found, using default\n", config.Theme)
		config.Theme = "Dracula"
		// Save corrected config
		SaveConfig(config)
	}

	// Validate avatar width (30-240)
	if config.AvatarWidth < 30 || config.AvatarWidth > 240 {
		if config.AvatarWidth != 0 { // 0 means not set, use default
			fmt.Fprintf(os.Stderr, "Warning: Avatar width %d out of range (30-240), using default\n", config.AvatarWidth)
		}
		config.AvatarWidth = 60
		SaveConfig(config)
	}

	// Validate avatar renderer
	if config.AvatarRenderer != "mosaic" && config.AvatarRenderer != "ascii" {
		if config.AvatarRenderer != "" {
			fmt.Fprintf(os.Stderr, "Warning: Avatar renderer '%s' invalid, using default\n", config.AvatarRenderer)
		}
		config.AvatarRenderer = "mosaic"
		SaveConfig(config)
	}

	return config, nil
}

// SaveConfig saves the configuration to ~/.local/gittui/config.json
func SaveConfig(config Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// ListAvailableThemes returns all available theme names (361 total)
func ListAvailableThemes() []string {
	return goghthemes.Names()
}
