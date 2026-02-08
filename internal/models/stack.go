package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

// StackConfig represents the configuration for a stack
type StackConfig struct {
	ID          string    `json:"id"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	Ignore      []string  `json:"ignore"`
}

// ParseStackConfig reads and parses boiler.stack.json from a directory
func ParseStackConfig(dirPath string) (*StackConfig, error) {
	configPath := filepath.Join(dirPath, "boiler.stack.json")
	if !utils.FileExists(configPath) {
		return nil, fmt.Errorf("Cannot find 'boiler.stack.json' in the current stack. Make sure you have initialized this stack by running 'bl init' inside the project folder")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read boiler.stack.json: %w", err)
	}

	var config StackConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse boiler.stack.json: %w", err)
	}

	return &config, nil
}

// ResolveIgnorePatterns returns ignore patterns from config
func ResolveIgnorePatterns(config *StackConfig) []string {
	// Use patterns from config, always add boiler.stack.json
	patterns := make([]string, len(config.Ignore))
	copy(patterns, config.Ignore)
	return append(patterns, "boiler.stack.json")
}
