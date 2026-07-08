package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigFile represents a physical configuration file on disk.
type ConfigFile struct {
	Path   string
	Exists bool
	Config *Config
}

// Manager orchestrates Global and Local ConfigFiles, and provides the merged Runtime Config.
type Manager struct {
	Global  *ConfigFile
	Local   *ConfigFile
	Runtime *Config
}

// BoilerContext holds the Manager and the resolved operational Scope for the current command.
type BoilerContext struct {
	Manager *Manager
	Scope   Scope
}

// SaveGlobal writes the Global ConfigFile's data back to disk.
func (m *Manager) SaveGlobal() error {
	if m.Global == nil {
		return fmt.Errorf("global config file is nil")
	}
	return m.Global.Save()
}

// SaveLocal writes the Local ConfigFile's data back to disk.
func (m *Manager) SaveLocal() error {
	if m.Local == nil {
		return fmt.Errorf("local config file is nil")
	}
	return m.Local.Save()
}

// Save writes the ConfigFile's data back to disk.
// It automatically creates the parent directory if it does not exist.
func (c *ConfigFile) Save() error {
	if c.Path == "" {
		return fmt.Errorf("cannot save config: path is empty")
	}

	// Ensure the parent directory exists
	dir := filepath.Dir(c.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(c.Config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.Path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file to %s: %w", c.Path, err)
	}

	c.Exists = true
	return nil
}
