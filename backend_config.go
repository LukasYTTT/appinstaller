package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DesktopIcon    bool   `json:"desktop_icon"`
	DesktopPath    string `json:"desktop_path"`
	NoExtractIcon  bool   `json:"no_extract_icon"`
	AutoLaunch     bool   `json:"auto_launch"`
}

func LoadConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "appimage-installer")
	configFile := filepath.Join(configDir, "config.json")

	// Defaults
	cfg := &Config{
		DesktopIcon:   true,
		DesktopPath:   filepath.Join(homeDir, "Schreibtisch"),
		NoExtractIcon: false,
		AutoLaunch:    false,
	}

	data, err := os.ReadFile(configFile)
	if err == nil {
		_ = json.Unmarshal(data, cfg)
	}

	return cfg
}

func (c *Config) Save() error {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "appimage-installer")
	configFile := filepath.Join(configDir, "config.json")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}
