package store

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// ── ParseResourceName ─────────────────────────────────────────────────────────

func TestParseResourceName(t *testing.T) {
	tests := []struct {
		input     string
		wantName  string
		wantVer   string
		wantExt   string
	}{
		// snippet with version and extension
		{"logger@1.js", "logger", "1", ".js"},
		{"errorHandler@3.ts", "errorHandler", "3", ".ts"},
		// snippet without version
		{"logger.js", "logger", "", ".js"},
		// stack (no extension)
		{"express-api@2", "express-api", "2", ""},
		// stack without version
		{"express-api", "express-api", "", ""},
		// extensionless file like Dockerfile
		{"Dockerfile", "Dockerfile", "", ""},
		// multi-dot extension
		{"config@1.test.js", "config@1.test", "", ".js"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, ver, ext := ParseResourceName(tt.input)
			if name != tt.wantName {
				t.Errorf("name: got %q, want %q", name, tt.wantName)
			}
			if ver != tt.wantVer {
				t.Errorf("version: got %q, want %q", ver, tt.wantVer)
			}
			if ext != tt.wantExt {
				t.Errorf("ext: got %q, want %q", ext, tt.wantExt)
			}
		})
	}
}

// ── IsStack / IsSnippet ───────────────────────────────────────────────────────

func TestIsStackIsSnippet(t *testing.T) {
	tests := []struct {
		input    string
		isStack  bool
		isSnippet bool
	}{
		{"express-api@1", true, false},
		{"logger@1.js", false, true},
		{"logger.js", false, true},
		{"express-api", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsStack(tt.input); got != tt.isStack {
				t.Errorf("IsStack(%q) = %v, want %v", tt.input, got, tt.isStack)
			}
			if got := IsSnippet(tt.input); got != tt.isSnippet {
				t.Errorf("IsSnippet(%q) = %v, want %v", tt.input, got, tt.isSnippet)
			}
		})
	}
}

// ── IsRemotePath ──────────────────────────────────────────────────────────────

func TestIsRemotePath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"owner/repo", true},
		{"owner/repo:path/file.js", true},
		{"https://github.com/owner/repo", true},
		{"http://example.com/file.js", true},
		{"iamabhinav.dev:snippets/validator.js", true},
		// local paths
		{"./local/path", false},
		{"/absolute/path", false},
		{"localfile.js", false},
		{"express-api@1", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsRemotePath(tt.input); got != tt.want {
				t.Errorf("IsRemotePath(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ── ParseRemotePath ───────────────────────────────────────────────────────────

func TestParseRemotePath(t *testing.T) {
	tests := []struct {
		input     string
		wantOwner string
		wantRepo  string
		wantPath  string
	}{
		{"owner/repo", "owner", "repo", "."},
		{"owner/repo:path/file.js", "owner", "repo", "path/file.js"},
		{"https://example.com/file.js", "", "https://example.com/file.js", ""},
		{"iamabhinav.dev:snippets/file.js", "", "iamabhinav.dev", "snippets/file.js"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			owner, repo, path := ParseRemotePath(tt.input)
			if owner != tt.wantOwner {
				t.Errorf("owner: got %q, want %q", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo: got %q, want %q", repo, tt.wantRepo)
			}
			if path != tt.wantPath {
				t.Errorf("path: got %q, want %q", path, tt.wantPath)
			}
		})
	}
}

// ── GetAllVersions ────────────────────────────────────────────────────────────

func TestGetAllVersions(t *testing.T) {
	dir := t.TempDir()
	st := NewStore(dir)
	// Pre-populate meta without touching disk
	st.meta.Snippets = map[string]string{
		"logger@1.js": "/path/logger@1.js",
		"logger@3.js": "/path/logger@3.js",
		"logger@5.js": "/path/logger@5.js",
		"other@1.js":  "/path/other@1.js",
	}

	got := st.GetAllVersions("logger", ".js")
	want := []int{1, 3, 5}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetAllVersions = %v, want %v", got, want)
	}

	// No versions
	got = st.GetAllVersions("nonexistent", ".js")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

// ── Store Load / Save / Add / Remove ─────────────────────────────────────────

func TestStoreLoadSave(t *testing.T) {
	dir := t.TempDir()
	st := NewStore(dir)

	if err := st.AddSnippet("logger@1.js", "/tmp/logger.js"); err != nil {
		t.Fatalf("AddSnippet: %v", err)
	}
	if err := st.AddStack("express@1", "/tmp/express"); err != nil {
		t.Fatalf("AddStack: %v", err)
	}

	// Reload from disk
	st2 := NewStore(dir)
	if err := st2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !st2.SnippetExists("logger@1.js") {
		t.Error("expected snippet to exist after reload")
	}
	if !st2.StackExists("express@1") {
		t.Error("expected stack to exist after reload")
	}

	// Remove and verify
	if err := st2.RemoveSnippet("logger@1.js"); err != nil {
		t.Fatalf("RemoveSnippet: %v", err)
	}
	if st2.SnippetExists("logger@1.js") {
		t.Error("expected snippet to be gone after remove")
	}
}

func TestStoreMetaFileCreatedOnMissing(t *testing.T) {
	dir := t.TempDir()
	metaPath := filepath.Join(dir, "boiler.meta.json")

	// File should not exist yet
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Fatal("meta file should not exist before Load")
	}

	st := NewStore(dir)
	if err := st.Load(); err != nil {
		t.Fatalf("Load on missing meta: %v", err)
	}

	// File should be created now
	if _, err := os.Stat(metaPath); err != nil {
		t.Errorf("meta file should be created after Load: %v", err)
	}
}
