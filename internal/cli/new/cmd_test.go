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

	got, err := resolveScriptPath("routes", localDir, globalDir, config.ScopeGlobal)
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

	got, err := resolveScriptPath("missing", tmp, tmp, config.ScopeGlobal)
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

	got, err := resolveScriptPath("routes", localDir, globalDir, config.ScopeLocal)
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

	got, err := resolveScriptPath("routes", localDir, globalDir, config.ScopeLocal)
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

	got, err := resolveScriptPath("missing", tmp, tmp, config.ScopeLocal)
	if err == nil {
		t.Fatalf("expected error, got path %q", got)
	}
}
