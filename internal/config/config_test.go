package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rishiyaduwanshi/boiler/internal/constants"
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

	expectedRoot := filepath.Join(home, constants.GlobalBoilerDirName)
	if cfg.Paths.Root != expectedRoot {
		t.Errorf("Paths.Root = %q, want %q", cfg.Paths.Root, expectedRoot)
	}
	if cfg.Paths.Snippets != filepath.Join(expectedRoot, constants.StoreDirName, constants.SnippetsDirName) {
		t.Errorf("Paths.Snippets = %q", cfg.Paths.Snippets)
	}
	if cfg.Paths.Stacks != filepath.Join(expectedRoot, constants.StoreDirName, constants.StacksDirName) {
		t.Errorf("Paths.Stacks = %q", cfg.Paths.Stacks)
	}
	if cfg.Paths.Logs != filepath.Join(expectedRoot, constants.LogsDirName) {
		t.Errorf("Paths.Logs = %q", cfg.Paths.Logs)
	}
	if len(cfg.Artifacts) == 0 {
		t.Error("Artifacts map should not be empty")
	}
	if cfg.Aliases == nil {
		t.Error("Aliases map should not be nil")
	}
	if cfg.Vars == nil {
		t.Error("Vars map should not be nil")
	}
}

// Save and Load logic is now tested via Manager in loader_test.go or manager_test.go

// ── mergeWithDefaults fills missing fields ────────────────────────────────────

func TestMergeWithDefaults(t *testing.T) {
	redirectHome(t)

	// Partial config - missing several fields
	partial := &Config{
		DefaultEditor: "emacs",
		// Registry, Name, Paths, Artifacts, Aliases, Vars all zero
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
	if partial.Vars == nil {
		t.Error("Vars should be initialized by mergeWithDefaults")
	}
}

func TestMergeWithDefaults_NormalizesVars(t *testing.T) {
	redirectHome(t)

	cfg := &Config{
		Vars: map[string]string{
			"bl__API_URL": "https://api.example.com",
			":TEAM_REG":   "https://github.com/myorg/boiler",
		},
	}

	mergeWithDefaults(cfg)

	if cfg.Vars["api_url"] != "https://api.example.com" {
		t.Fatalf("expected normalized key api_url")
	}
	if cfg.Vars["team_reg"] != "https://github.com/myorg/boiler" {
		t.Fatalf("expected normalized key team_reg")
	}
}

func TestMergeWithDefaults_ArtifactsPartial(t *testing.T) {
	redirectHome(t)

	// Config with some artifacts - merge should add missing ones without removing existing
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

// Backup and Reset logic will be handled by Manager in the future.

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

// ── JSON serialization preserves all fields ───────────────────────────────────

func TestConfigJSONRoundtrip(t *testing.T) {
	redirectHome(t)

	cfg := DefaultConfig()
	cfg.Aliases["ll"] = "list --long"
	cfg.Vars["api_url"] = "https://api.example.com"

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
	if restored.Vars["api_url"] != "https://api.example.com" {
		t.Error("vars not preserved across JSON roundtrip")
	}
	if restored.Paths.Snippets != cfg.Paths.Snippets {
		t.Error("Paths.Snippets not preserved across JSON roundtrip")
	}
}

func TestLoad_UppercaseAliasAutoNormalized(t *testing.T) {
	tmp := redirectHome(t)

	globalDir := filepath.Join(tmp, ".boiler")
	os.MkdirAll(globalDir, 0755)
	globalPath := filepath.Join(globalDir, "boiler.conf.json")
	mixedCaseJSON := `{"aliases": {"LL": "ls"}}`
	os.WriteFile(globalPath, []byte(mixedCaseJSON), 0644)

	mgr, err := Load()
	if err != nil {
		t.Fatalf("Load() should not fail for mixed-case alias 'LL': %v", err)
	}

	if v, ok := mgr.Global.Config.Aliases["ll"]; !ok || v != "ls" {
		t.Errorf("expected 'LL' to be auto-normalized to 'll', got map: %v", mgr.Global.Config.Aliases)
	}
}
