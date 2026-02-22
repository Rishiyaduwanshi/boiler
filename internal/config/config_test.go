package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// redirectHome points os.UserHomeDir() at a temp dir for the duration of the test.
func redirectHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp) // Windows
	return tmp
}

// ── DefaultConfig ─────────────────────────────────────────────────────────────

func TestDefaultConfig(t *testing.T) {
	home := redirectHome(t)
	cfg := DefaultConfig()

	if cfg.Name == "" {
		t.Error("Name should not be empty")
	}
	if cfg.Registry == "" {
		t.Error("Registry should not be empty")
	}
	if cfg.DefaultEditor == "" {
		t.Error("DefaultEditor should not be empty")
	}
	if cfg.Version == "" {
		t.Error("Version should not be empty")
	}

	expectedRoot := filepath.Join(home, ".boiler")
	if cfg.Paths.Root != expectedRoot {
		t.Errorf("Paths.Root = %q, want %q", cfg.Paths.Root, expectedRoot)
	}
	if cfg.Paths.Snippets != filepath.Join(expectedRoot, "store", "snippets") {
		t.Errorf("Paths.Snippets = %q", cfg.Paths.Snippets)
	}
	if cfg.Paths.Stacks != filepath.Join(expectedRoot, "store", "stacks") {
		t.Errorf("Paths.Stacks = %q", cfg.Paths.Stacks)
	}
	if cfg.Paths.Logs != filepath.Join(expectedRoot, "logs") {
		t.Errorf("Paths.Logs = %q", cfg.Paths.Logs)
	}
	if len(cfg.Artifacts) == 0 {
		t.Error("Artifacts map should not be empty")
	}
	if cfg.Aliases == nil {
		t.Error("Aliases map should not be nil")
	}
}

// ── Save / Load roundtrip ─────────────────────────────────────────────────────

func TestSaveAndLoad(t *testing.T) {
	redirectHome(t)

	original := DefaultConfig()
	original.DefaultEditor = "nano"
	original.Registry = "https://github.com/test/repo"

	if err := Save(original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.DefaultEditor != "nano" {
		t.Errorf("DefaultEditor: got %q, want %q", loaded.DefaultEditor, "nano")
	}
	if loaded.Registry != "https://github.com/test/repo" {
		t.Errorf("Registry: got %q, want %q", loaded.Registry, "https://github.com/test/repo")
	}
	if loaded.Paths.Root != original.Paths.Root {
		t.Errorf("Paths.Root changed after reload")
	}
}

// ── Load on missing file creates default ─────────────────────────────────────

func TestLoadCreatesDefaultOnMissing(t *testing.T) {
	redirectHome(t)

	// No config file exists yet
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Name == "" {
		t.Error("expected default name to be set")
	}

	// Config file should now exist on disk
	configPath, _ := ConfigPath()
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("config file should have been created: %v", err)
	}
}

// ── mergeWithDefaults fills missing fields ────────────────────────────────────

func TestMergeWithDefaults(t *testing.T) {
	redirectHome(t)

	// Partial config — missing several fields
	partial := &Config{
		DefaultEditor: "emacs",
		// Registry, Name, Paths, Artifacts, Aliases all zero
	}
	mergeWithDefaults(partial)

	if partial.DefaultEditor != "emacs" {
		t.Error("mergeWithDefaults should NOT overwrite set fields")
	}
	if partial.Registry == "" {
		t.Error("Registry should be filled by mergeWithDefaults")
	}
	if partial.Name == "" {
		t.Error("Name should be filled by mergeWithDefaults")
	}
	if partial.Paths.Root == "" {
		t.Error("Paths.Root should be filled by mergeWithDefaults")
	}
	if len(partial.Artifacts) == 0 {
		t.Error("Artifacts should be filled by mergeWithDefaults")
	}
	if partial.Aliases == nil {
		t.Error("Aliases should be initialized by mergeWithDefaults")
	}
}

func TestMergeWithDefaults_ArtifactsPartial(t *testing.T) {
	redirectHome(t)

	// Config with some artifacts — merge should add missing ones without removing existing
	cfg := &Config{
		Artifacts: map[string]string{
			"custom": "## ",
		},
	}
	mergeWithDefaults(cfg)

	if cfg.Artifacts["custom"] != "## " {
		t.Error("custom artifact should be preserved")
	}
	if _, ok := cfg.Artifacts["go"]; !ok {
		t.Error("go artifact should be added by mergeWithDefaults")
	}
	if _, ok := cfg.Artifacts["py"]; !ok {
		t.Error("py artifact should be added by mergeWithDefaults")
	}
}

// ── CreateBackup + Reset ──────────────────────────────────────────────────────

func TestCreateBackupAndResetFromBackup(t *testing.T) {
	redirectHome(t)

	// Save an initial config
	original := DefaultConfig()
	original.DefaultEditor = "vim-original"
	if err := Save(original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Create backup
	if err := CreateBackup(); err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	backupPath, _ := BackupPath()
	if _, err := os.Stat(backupPath); err != nil {
		t.Fatalf("backup file should exist: %v", err)
	}

	// Overwrite config with different value
	modified := DefaultConfig()
	modified.DefaultEditor = "nvim-modified"
	if err := Save(modified); err != nil {
		t.Fatalf("Save modified: %v", err)
	}

	// Reset should restore from backup
	if err := Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	restored, err := Load()
	if err != nil {
		t.Fatalf("Load after reset: %v", err)
	}
	if restored.DefaultEditor != "vim-original" {
		t.Errorf("Reset should restore from backup; got editor %q", restored.DefaultEditor)
	}
}

func TestResetWithNoBackup(t *testing.T) {
	redirectHome(t)

	// No backup file — Reset should write default config
	if err := Reset(); err != nil {
		t.Fatalf("Reset with no backup: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load after reset: %v", err)
	}
	defaults := DefaultConfig()
	if cfg.Name != defaults.Name {
		t.Errorf("Name after reset = %q, want %q", cfg.Name, defaults.Name)
	}
}

// ── CreateBackup on missing config ───────────────────────────────────────────

func TestCreateBackup_NoConfigFile(t *testing.T) {
	redirectHome(t)
	// Should not error when source doesn't exist — just a no-op
	if err := CreateBackup(); err != nil {
		t.Errorf("CreateBackup with no config should not error: %v", err)
	}
}

// ── InitializeDirs ────────────────────────────────────────────────────────────

func TestInitializeDirs(t *testing.T) {
	redirectHome(t)

	cfg := DefaultConfig()
	if err := cfg.InitializeDirs(); err != nil {
		t.Fatalf("InitializeDirs: %v", err)
	}

	dirs := []string{
		cfg.Paths.Root,
		cfg.Paths.Store,
		cfg.Paths.Snippets,
		cfg.Paths.Stacks,
		cfg.Paths.Logs,
		cfg.Paths.Bin,
	}
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("dir %q not created: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q should be a directory", dir)
		}
	}
}

// ── ConfigPath / BackupPath use home dir ──────────────────────────────────────

func TestConfigPath(t *testing.T) {
	home := redirectHome(t)
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}
	expected := filepath.Join(home, ".boiler", "boiler.conf.json")
	if path != expected {
		t.Errorf("ConfigPath = %q, want %q", path, expected)
	}
}

// ── JSON serialization preserves all fields ───────────────────────────────────

func TestConfigJSONRoundtrip(t *testing.T) {
	redirectHome(t)

	cfg := DefaultConfig()
	cfg.Aliases["ll"] = "list --long"

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var restored Config
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if restored.Aliases["ll"] != "list --long" {
		t.Error("alias not preserved across JSON roundtrip")
	}
	if restored.Paths.Snippets != cfg.Paths.Snippets {
		t.Error("Paths.Snippets not preserved across JSON roundtrip")
	}
}
