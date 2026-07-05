package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/rishiyaduwanshi/boiler/pkg/version"
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
	Vars          map[string]string `json:"vars"`
}

type Paths struct {
	Root     string `json:"root"`
	Store    string `json:"store"`
	Snippets string `json:"snippets"`
	Stacks   string `json:"stacks"`
	Logs     string `json:"logs"`
	Bin      string `json:"bin"`
}

// defaultEditor returns the OS-appropriate default editor
func defaultEditor() string {
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "vi"
}

// getRootPath returns the default root directory for boiler
func getRootPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".boiler"
	}
	return filepath.Join(home, ".boiler")
}

func DefaultConfig() *Config {
	rootPath := getRootPath()

	ver := version.Version
	if ver == "" {
		ver = "dev"
	}

	return &Config{
		Name:          "Boiler",
		Version:       ver,
		Author:        "Abhinav Prakash",
		Github:        "github.com/rishiyaduwanshi/boiler",
		Description:   "A CLI tool to manage reusable code snippets and stacks",
		DefaultEditor: defaultEditor(),
		Registry:      "https://github.com/rishiyaduwanshi/boiler",
		Paths: Paths{
			Root:     rootPath,
			Store:    filepath.Join(rootPath, "store"),
			Snippets: filepath.Join(rootPath, "store", "snippets"),
			Stacks:   filepath.Join(rootPath, "store", "stacks"),
			Logs:     filepath.Join(rootPath, "logs"),
			Bin:      filepath.Join(rootPath, "bin"),
		},
		Artifacts: map[string]string{
			"default":    "//  ",
			"bl":         "//  ",
			"js":         "//  ",
			"ts":         "//  ",
			"jsx":        "//  ",
			"tsx":        "//  ",
			"java":       "//  ",
			"c":          "//  ",
			"cpp":        "//  ",
			"go":         "//  ",
			"rs":         "//  ",
			"py":         "#  ",
			"rb":         "#  ",
			"sh":         "#  ",
			"bash":       "#  ",
			"ps1":        "#  ",
			"html":       "<!--  -->",
			"htm":        "<!--  -->",
			"css":        "/*  */",
			"sql":        "--  ",
			"yml":        "#  ",
			"yaml":       "#  ",
			"xml":        "<!--  -->",
			"md":         "<!--  -->",
			"ahk":        ";  ",
			"dockerfile": "#  ",
			"gitignore":  "#  ",
			"env":        "#  ",
			"toml":       "#  ",
			"ini":        ";  ",
		},
		Aliases: map[string]string{
			"gi": "add github/gitignore:bl__1.gitignore . -m .gitignore -r",
		},
		Vars: make(map[string]string),
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	rootPath := getRootPath()
	return filepath.Join(rootPath, "boiler.conf.json"), nil
}

func BackupPath() (string, error) {
	rootPath := getRootPath()
	return filepath.Join(rootPath, "boiler.conf.json.bk"), nil
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

	// Fill in any missing fields with defaults (handles config upgrades)
	mergeWithDefaults(&cfg)

	return &cfg, nil
}

// mergeWithDefaults fills zero-value fields in cfg with values from DefaultConfig.
// This ensures existing users get new fields after a Boiler upgrade.
func mergeWithDefaults(cfg *Config) {
	defaults := DefaultConfig()

	if cfg.DefaultEditor == "" {
		cfg.DefaultEditor = defaults.DefaultEditor
	}
	if cfg.Registry == "" {
		cfg.Registry = defaults.Registry
	}
	if cfg.Name == "" {
		cfg.Name = defaults.Name
	}
	if cfg.Author == "" {
		cfg.Author = defaults.Author
	}
	if cfg.Github == "" {
		cfg.Github = defaults.Github
	}
	if cfg.Version == "" {
		cfg.Version = defaults.Version
	}

	// Ensure paths are populated (in case root changed)
	if cfg.Paths.Root == "" {
		cfg.Paths = defaults.Paths
	}

	// Only add artifact keys that the user hasn't set.
	// User's existing values are always respected.
	if cfg.Artifacts == nil {
		cfg.Artifacts = defaults.Artifacts
	} else {
		for k, v := range defaults.Artifacts {
			if _, exists := cfg.Artifacts[k]; !exists {
				cfg.Artifacts[k] = v
			}
		}
	}

	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
	}

	if cfg.Vars == nil {
		cfg.Vars = make(map[string]string)
	} else {
		cfg.Vars = utils.NormalizeVarMap(cfg.Vars)
	}
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
