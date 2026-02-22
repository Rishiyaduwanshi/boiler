package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── shouldIgnore ──────────────────────────────────────────────────────────────

func TestShouldIgnore(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		want     bool
	}{
		// exact match
		{"node_modules", []string{"node_modules", ".git"}, true},
		{".git", []string{"node_modules", ".git"}, true},
		// not in list
		{"src", []string{"node_modules", ".git"}, false},
		// glob: *.log
		{"app.log", []string{"*.log"}, true},
		{"debug.log", []string{"*.log"}, true},
		{"app.txt", []string{"*.log"}, false},
		// glob: *.tmp
		{"cache.tmp", []string{"*.tmp", "*.log"}, true},
		// empty patterns
		{"anything", []string{}, false},
		// .DS_Store exact
		{".DS_Store", DefaultIgnorePatterns, true},
		// Thumbs.db exact
		{"Thumbs.db", DefaultIgnorePatterns, true},
	}

	for _, tt := range tests {
		tt := tt
		label := tt.name
		if len(tt.patterns) > 0 {
			label += "_" + tt.patterns[0]
		}
		t.Run(label, func(t *testing.T) {
			if got := shouldIgnore(tt.name, tt.patterns); got != tt.want {
				t.Errorf("shouldIgnore(%q, %v) = %v, want %v", tt.name, tt.patterns, got, tt.want)
			}
		})
	}
}

// ── FileExists ────────────────────────────────────────────────────────────────

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(existing, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	if !FileExists(existing) {
		t.Errorf("FileExists(%q) = false, want true", existing)
	}
	if FileExists(filepath.Join(dir, "nonexistent.txt")) {
		t.Error("FileExists on missing path = true, want false")
	}
}

// ── IsDirectory ───────────────────────────────────────────────────────────────

func TestIsDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	if !IsDirectory(dir) {
		t.Errorf("IsDirectory(%q) = false, want true", dir)
	}
	if IsDirectory(file) {
		t.Errorf("IsDirectory(%q) = true on file, want false", file)
	}
	if IsDirectory(filepath.Join(dir, "missing")) {
		t.Error("IsDirectory on missing path = true, want false")
	}
}

// ── CopyFile ──────────────────────────────────────────────────────────────────

func TestCopyFile(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	srcFile := filepath.Join(src, "hello.txt")
	content := []byte("hello world\n")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	dstFile := filepath.Join(dst, "sub", "hello.txt")
	if err := CopyFile(srcFile, dstFile); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}

	got, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("reading dst: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}

func TestCopyFile_MissingSource(t *testing.T) {
	err := CopyFile("/nonexistent/src.txt", t.TempDir()+"/dst.txt")
	if err == nil {
		t.Error("expected error for missing source, got nil")
	}
}

// ── CopyDir ───────────────────────────────────────────────────────────────────

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create tree: src/a.txt, src/sub/b.txt, src/node_modules/dep.js
	write := func(rel, content string) {
		path := filepath.Join(src, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("a.txt", "file a")
	write("sub/b.txt", "file b")
	write("node_modules/dep.js", "should be ignored")

	if err := CopyDir(src, dst, DefaultIgnorePatterns); err != nil {
		t.Fatalf("CopyDir: %v", err)
	}

	// a.txt and sub/b.txt should exist
	for _, rel := range []string{"a.txt", "sub/b.txt"} {
		if !FileExists(filepath.Join(dst, rel)) {
			t.Errorf("expected %s to be copied", rel)
		}
	}
	// node_modules should be ignored
	if FileExists(filepath.Join(dst, "node_modules", "dep.js")) {
		t.Error("node_modules should have been ignored")
	}
}

// ── CopyFileWithVariables ────────────────────────────────────────────────────

func TestCopyFileWithVariables(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "snippet.js")
	dst := filepath.Join(dir, "out.js")

	// File with metadata comments and template vars
	srcContent := `// __author rishiyaduwanshi
// __desc A sample snippet
const url = "bl__API_URL";
const key = "bl__API_KEY";
console.log("ready");
`
	if err := os.WriteFile(src, []byte(srcContent), 0644); err != nil {
		t.Fatal(err)
	}

	replacements := map[string]string{
		"bl__API_URL": "https://api.example.com",
		"bl__API_KEY": "abc123",
	}
	if err := CopyFileWithVariables(src, dst, replacements); err != nil {
		t.Fatalf("CopyFileWithVariables: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)

	// metadata lines should be stripped
	if strings.Contains(out, "__author") || strings.Contains(out, "__desc") {
		t.Error("metadata lines should have been removed")
	}
	// variables should be replaced
	if strings.Contains(out, "bl__API_URL") || strings.Contains(out, "bl__API_KEY") {
		t.Error("template variables should have been replaced")
	}
	if !strings.Contains(out, "https://api.example.com") {
		t.Error("expected replaced URL value in output")
	}
	if !strings.Contains(out, "abc123") {
		t.Error("expected replaced key value in output")
	}
	// non-metadata code lines should survive
	if !strings.Contains(out, `console.log("ready")`) {
		t.Error("regular code line should be preserved")
	}
}

func TestCopyFileWithVariables_NilReplacements(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "plain.js")
	dst := filepath.Join(dir, "out.js")

	if err := os.WriteFile(src, []byte("const x = 1;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := CopyFileWithVariables(src, dst, nil); err != nil {
		t.Fatalf("CopyFileWithVariables with nil replacements: %v", err)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "const x = 1;\n" {
		t.Errorf("content changed unexpectedly: %q", got)
	}
}

