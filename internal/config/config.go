package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Author        string            `json:"author"`
	Github        string            `json:"github"`
	Description   string            `json:"description"`
	DefaultEditor string            `json:"defaultEditor"`
	Registry      string            `json:"registry"`
	Paths         Paths             `json:"paths"`
	Artifacts     map[string]string `json:"artifacts"`
	Aliases       map[string]string `json:"aliases"`
}

type Paths struct {
	Root     string `json:"root"`
	Store    string `json:"store"`
	Snippets string `json:"snippets"`
	Stacks   string `json:"stacks"`
	Logs     string `json:"logs"`
	Bin      string `json:"bin"`
}

func DefaultConfig() *Config {
	return &Config{
		Name:          "Boiler",
		Version:       "0.0.1",
		Author:        "Abhinav Prakash",
		Github:        "github.com/rishiyaduwanshi/boiler",
		Description:   "A CLI tool to manage reusable code snippets and stacks",
		DefaultEditor: "vim",
		Registry:      "https://github.com/rishiyaduwanshi/boiler/store",
		Paths: Paths{
			Root:     "~/.boiler",
			Store:    "~/.boiler/store",
			Snippets: "~/.boiler/store/snippets",
			Stacks:   "~/.boiler/store/stacks",
			Logs:     "~/.boiler/logs",
			Bin:      "~/.boiler/bin",
		},
		Artifacts: map[string]string{
			"default":    "//  ",
			"py":         "#  ",
			"rb":         "#  ",
			"sh":         "#  ",
			"bash":       "#  ",
			"ps1":        "#  ",
			"html":       "<!--  ",
			"htm":        "<!--  ",
			"css":        "/*  ",
			"sql":        "--  ",
			"yml":        "#  ",
			"yaml":       "#  ",
			"xml":        "<!--  ",
			"md":         "<!--  ",
			"ahk":        ";  ",
			"dockerfile": "#  ",
			"gitignore":  "#  ",
			"env":        "#  ",
			"toml":       "#  ",
			"ini":        ";  ",
		},
		Aliases: make(map[string]string),
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".boiler", "boiler.conf.json"), nil
}

func BackupPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".boiler", "boiler.conf.json.bk"), nil
}

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	return os.ExpandEnv(path)
}

func (p *Paths) ExpandPaths() {
	p.Root = ExpandPath(p.Root)
	p.Store = ExpandPath(p.Store)
	p.Snippets = ExpandPath(p.Snippets)
	p.Stacks = ExpandPath(p.Stacks)
	p.Logs = ExpandPath(p.Logs)
	p.Bin = ExpandPath(p.Bin)
}

func Load() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			if err := Save(cfg); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.Paths.ExpandPaths()

	return &cfg, nil
}

func Save(cfg *Config) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func Reset() error {
	backupPath, err := BackupPath()
	if err != nil {
		return err
	}

	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(backupPath); err == nil {
		data, err := os.ReadFile(backupPath)
		if err != nil {
			return fmt.Errorf("failed to read backup: %w", err)
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
		return nil
	}

	return Save(DefaultConfig())
}

func CreateBackup() error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	backupPath, err := BackupPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	return nil
}

func (cfg *Config) InitializeDirs() error {
	dirs := []string{
		cfg.Paths.Root,
		cfg.Paths.Store,
		cfg.Paths.Snippets,
		cfg.Paths.Stacks,
		cfg.Paths.Logs,
		cfg.Paths.Bin,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
