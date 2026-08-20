package new

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rishiyaduwanshi/boiler/internal/config"
)

// makeScript creates a temporary .bl file at the given path for testing.
func makeScript(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("makeScript: mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte("// test script"), 0644); err != nil {
		t.Fatalf("makeScript: write %s: %v", path, err)
	}
}

func TestResolveScriptPath_GlobalScope(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	globalDir := filepath.Join(tmp, "global", "commands")
	localDir := filepath.Join(tmp, "local", "bl", "commands")

	makeScript(t, filepath.Join(globalDir, "routes.bl"))

	got, err := resolveScriptPath("routes", localDir, globalDir, config.ScopeGlobal, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(globalDir, "routes.bl")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveScriptPath_GlobalScope_NotFound(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	got, err := resolveScriptPath("missing", tmp, tmp, config.ScopeGlobal, false)
	if err == nil {
		t.Fatalf("expected error, got path %q", got)
	}
}

func TestResolveScriptPath_LocalScope_UsesLocalFirst(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	globalDir := filepath.Join(tmp, "global", "commands")
	localDir := filepath.Join(tmp, "local", "bl", "commands")

	// Both exist; local should win.
	makeScript(t, filepath.Join(globalDir, "routes.bl"))
	makeScript(t, filepath.Join(localDir, "routes.bl"))

	got, err := resolveScriptPath("routes", localDir, globalDir, config.ScopeLocal, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(localDir, "routes.bl")
	if got != want {
		t.Fatalf("got %q, want %q (expected local to win)", got, want)
	}
}

func TestResolveScriptPath_LocalScope_FallsBackToGlobal(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	globalDir := filepath.Join(tmp, "global", "commands")
	localDir := filepath.Join(tmp, "local", "bl", "commands")

	// Only global exists.
	makeScript(t, filepath.Join(globalDir, "routes.bl"))

	got, err := resolveScriptPath("routes", localDir, globalDir, config.ScopeLocal, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(globalDir, "routes.bl")
	if got != want {
		t.Fatalf("got %q, want %q (expected global fallback)", got, want)
	}
}

func TestResolveScriptPath_LocalScope_NotFound(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	got, err := resolveScriptPath("missing", tmp, tmp, config.ScopeLocal, false)
	if err == nil {
		t.Fatalf("expected error, got path %q", got)
	}
}

func TestResolveScriptPath_LocalScope_ExplicitLocal_NoFallback(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	globalDir := filepath.Join(tmp, "global", "commands")
	localDir := filepath.Join(tmp, "local", "bl", "commands")

	// Only global exists.
	makeScript(t, filepath.Join(globalDir, "routes.bl"))

	got, err := resolveScriptPath("routes", localDir, globalDir, config.ScopeLocal, true)
	if err == nil {
		t.Fatalf("expected error (no fallback for explicit local), got path %q", got)
	}
}

func TestCmdRun_GlobalOverride(t *testing.T) {
	// Setup temporary workspace and mock environments
	tmp := t.TempDir()

	// Create mock config to ensure FindNearestConfig is triggered
	localConfig := filepath.Join(tmp, "boiler.local.json")
	if err := os.WriteFile(localConfig, []byte(`{"scope": "local"}`), 0644); err != nil {
		t.Fatalf("failed to create local config: %v", err)
	}

	// Create local script (so it exists)
	localScript := filepath.Join(tmp, "bl", "commands", "testcmd.bl")
	makeScript(t, localScript)

	// Switch working directory to mock project root
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to switch to temp dir: %v", err)
	}
	defer func() {
		_ = os.Chdir(origDir)
	}()

	// Mock HOME/USERPROFILE env vars so config.DefaultConfig() points to tmp
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	// Mock injected CLI context config
	mockCfg := config.DefaultConfig()
	mockCfg.Paths.Commands = filepath.Join(tmp, "bl", "commands") // Match what root.go does for local scope

	Setup(mockCfg, nil)
	config.Ctx = &config.BoilerContext{
		Manager: &config.Manager{
			Runtime: mockCfg,
		},
		Scope: config.ScopeLocal,
	}

	// Execute command with --global override.
	// Since the global script does NOT exist (~/.boiler/commands/testcmd.bl):
	// - If fixed: resolves to global path, returns error because file is not found. (Test PASS)
	// - If bug: ignores --global, resolves to local script which exists, and runs successfully (returns nil). (Test FAIL)
	err = Cmd.RunE(Cmd, []string{"testcmd", "--global"})
	if err == nil {
		t.Fatalf("Bug present: expected 'not found' error because --global should bypass local script, but got nil (it resolved to local script!)")
	}
}

func TestCmdRun_LocalOverride_NoFallback(t *testing.T) {
	tmp := t.TempDir()

	localConfig := filepath.Join(tmp, "boiler.local.json")
	if err := os.WriteFile(localConfig, []byte(`{"scope": "local"}`), 0644); err != nil {
		t.Fatalf("failed to create local config: %v", err)
	}

	// Create global script (so it exists)
	globalScript := filepath.Join(tmp, "global", "commands", "testcmd.bl")
	makeScript(t, globalScript)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to switch to temp dir: %v", err)
	}
	defer func() {
		_ = os.Chdir(origDir)
	}()

	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	mockCfg := config.DefaultConfig()
	mockCfg.Paths.Commands = filepath.Join(tmp, "global", "commands")

	Setup(mockCfg, nil)
	config.Ctx = &config.BoilerContext{
		Manager: &config.Manager{
			Runtime: mockCfg,
		},
		Scope: config.ScopeLocal,
	}

	err = Cmd.RunE(Cmd, []string{"testcmd", "--local"})
	if err == nil {
		t.Fatalf("expected error because --local should prevent fallback to global script, but got nil")
	}
}
