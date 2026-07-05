package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

func TestResolveAddDestination(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPath string
	}{
		{name: "default destination", input: "", wantPath: addDefaultDestination},
		{name: "current directory", input: ".", wantPath: "."},
		{name: "relative path", input: "./src/utils", wantPath: filepath.Clean("./src/utils")},
		{name: "parent directory", input: "..", wantPath: ".."},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := resolveAddDestination(tt.input)
			if got != tt.wantPath {
				t.Fatalf("resolveAddDestination(%q) = %q, want %q", tt.input, got, tt.wantPath)
			}
		})
	}
}

func TestStackDirectoryName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "stack with version", in: "express@1", want: "express"},
		{name: "stack without version", in: "express", want: "express"},
		{name: "snippet keeps base name", in: "logger@2.js", want: "logger"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := stackDirectoryName(tt.in); got != tt.want {
				t.Fatalf("stackDirectoryName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCopyStackToDestination_NonSpread(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "app.js"), []byte("console.log('ok')\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dest := t.TempDir()
	opts := addOptions{spread: false, force: false}
	finalDest, err := copyStackToDestination(src, "express@1", dest, opts)
	if err != nil {
		t.Fatalf("copyStackToDestination returned error: %v", err)
	}

	wantDest := filepath.Join(dest, "express")
	if finalDest != wantDest {
		t.Fatalf("final destination = %q, want %q", finalDest, wantDest)
	}

	if _, err := os.Stat(filepath.Join(wantDest, "app.js")); err != nil {
		t.Fatalf("expected file copied to stack folder: %v", err)
	}
}

func TestCopyStackToDestination_Spread(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "server.js"), []byte("console.log('spread')\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dest := t.TempDir()
	opts := addOptions{spread: true, force: false}
	finalDest, err := copyStackToDestination(src, "express@1", dest, opts)
	if err != nil {
		t.Fatalf("copyStackToDestination returned error: %v", err)
	}

	if finalDest != dest {
		t.Fatalf("final destination = %q, want %q", finalDest, dest)
	}

	if _, err := os.Stat(filepath.Join(dest, "server.js")); err != nil {
		t.Fatalf("expected file copied into destination root: %v", err)
	}
}

func TestValidateSpreadDestination_Conflict(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "server.js"), []byte("src\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dest := t.TempDir()
	if err := os.WriteFile(filepath.Join(dest, "server.js"), []byte("dest\n"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := addOptions{force: false}
	err := validateSpreadDestination(src, dest, nil, opts)
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
}

func TestValidateSpreadDestination_IgnoresIgnoredEntries(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	if err := os.Mkdir(filepath.Join(src, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	dest := t.TempDir()
	if err := os.Mkdir(filepath.Join(dest, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	opts := addOptions{force: false}
	err := validateSpreadDestination(src, dest, utils.DefaultIgnorePatterns, opts)
	if err != nil {
		t.Fatalf("expected no error for ignored entry conflict, got: %v", err)
	}
}

func TestIsDirectRemoteFileURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "github blob file",
			input: "https://github.com/rishiyaduwanshi/boiler-snippets/blob/main/js/errorHandler.js",
			want:  true,
		},
		{
			name:  "github blob directory",
			input: "https://github.com/Rich-Harris/degit/blob/master/src",
			want:  false,
		},
		{
			name:  "github tree directory",
			input: "https://github.com/spf13/cobra/tree/main/doc",
			want:  false,
		},
		{
			name:  "raw github file",
			input: "https://raw.githubusercontent.com/rishiyaduwanshi/boiler-snippets/main/js/errorHandler.js",
			want:  true,
		},
		{
			name:  "direct archive url",
			input: "https://example.com/templates/express.zip",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isDirectRemoteFileURL(tt.input)
			if got != tt.want {
				t.Fatalf("isDirectRemoteFileURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveAddResourceType(t *testing.T) {
	tests := []struct {
		name        string
		stackFlag   bool
		snippetFlag bool
		want        addResourceType
		wantErr     bool
	}{
		{name: "auto by default", stackFlag: false, snippetFlag: false, want: addResourceTypeAuto, wantErr: false},
		{name: "force stack", stackFlag: true, snippetFlag: false, want: addResourceTypeStack, wantErr: false},
		{name: "force snippet", stackFlag: false, snippetFlag: true, want: addResourceTypeSnippet, wantErr: false},
		{name: "mutually exclusive flags", stackFlag: true, snippetFlag: true, want: addResourceTypeAuto, wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opts := addOptions{asStack: tt.stackFlag, asSnippet: tt.snippetFlag}
			got, err := resolveAddResourceType(opts)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("resolveAddResourceType() returned error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("resolveAddResourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldTreatDirectRemotePathAsSnippet(t *testing.T) {
	tests := []struct {
		name         string
		subPath      string
		resourceType addResourceType
		want         bool
	}{
		{name: "auto file path", subPath: "js/errorHandler.js", resourceType: addResourceTypeAuto, want: true},
		{name: "auto directory path", subPath: "templates", resourceType: addResourceTypeAuto, want: false},
		{name: "force snippet", subPath: "templates", resourceType: addResourceTypeSnippet, want: true},
		{name: "force stack", subPath: "js/errorHandler.js", resourceType: addResourceTypeStack, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := shouldTreatDirectRemotePathAsSnippet(tt.subPath, tt.resourceType)
			if got != tt.want {
				t.Fatalf("shouldTreatDirectRemotePathAsSnippet(%q, %v) = %v, want %v", tt.subPath, tt.resourceType, got, tt.want)
			}
		})
	}
}

func TestShouldTreatDirectURLAsSnippet(t *testing.T) {
	tests := []struct {
		name         string
		resourceURL  string
		resourceType addResourceType
		want         bool
	}{
		{name: "auto file url", resourceURL: "https://example.com/snippets/helper.js", resourceType: addResourceTypeAuto, want: true},
		{name: "auto archive url", resourceURL: "https://example.com/templates/express.zip", resourceType: addResourceTypeAuto, want: false},
		{name: "force snippet", resourceURL: "https://example.com/templates/noext", resourceType: addResourceTypeSnippet, want: true},
		{name: "force stack", resourceURL: "https://example.com/snippets/helper.js", resourceType: addResourceTypeStack, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := shouldTreatDirectURLAsSnippet(tt.resourceURL, tt.resourceType)
			if got != tt.want {
				t.Fatalf("shouldTreatDirectURLAsSnippet(%q, %v) = %v, want %v", tt.resourceURL, tt.resourceType, got, tt.want)
			}
		})
	}
}
