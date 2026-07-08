package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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
	Scope         string            `json:"scope,omitempty"`
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

// Backup functionality should be re-implemented in the Manager if needed.

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
