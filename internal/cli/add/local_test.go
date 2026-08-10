package add

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

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
	opts := Options{Spread: false, Force: false}
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
	opts := Options{Spread: true, Force: false}
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

func TestValidateStackDestination_ExistsWithoutForce(t *testing.T) {
	t.Parallel()

	destRoot := t.TempDir()
	// Non-spread places the stack in destRoot/<stackDir>
	conflict := filepath.Join(destRoot, "my-stack")
	if err := os.Mkdir(conflict, 0755); err != nil {
		t.Fatal(err)
	}

	err := validateStackDestination("alice/my-stack", destRoot, Options{Force: false})
	if err == nil {
		t.Fatal("expected destination-exists error, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error %q should mention already exists", err)
	}
}

func TestValidateStackDestination_ForceSkipsCheck(t *testing.T) {
	t.Parallel()

	destRoot := t.TempDir()
	conflict := filepath.Join(destRoot, "my-stack")
	if err := os.Mkdir(conflict, 0755); err != nil {
		t.Fatal(err)
	}

	if err := validateStackDestination("alice/my-stack", destRoot, Options{Force: true}); err != nil {
		t.Fatalf("force should skip existence check, got: %v", err)
	}
}

func TestValidateStackDestination_Available(t *testing.T) {
	t.Parallel()

	destRoot := t.TempDir()
	if err := validateStackDestination("alice/my-stack", destRoot, Options{}); err != nil {
		t.Fatalf("empty dest should be available, got: %v", err)
	}
}

func TestValidateStackDestination_CustomName(t *testing.T) {
	t.Parallel()

	destRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(destRoot, "custom"), 0755); err != nil {
		t.Fatal(err)
	}
	err := validateStackDestination("alice/my-stack", destRoot, Options{Name: "custom"})
	if err == nil {
		t.Fatal("expected conflict on custom name")
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

	opts := Options{Force: false}
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

	opts := Options{Force: false}
	err := validateSpreadDestination(src, dest, utils.DefaultIgnorePatterns, opts)
	if err != nil {
		t.Fatalf("expected no error for ignored entry conflict, got: %v", err)
	}
}
