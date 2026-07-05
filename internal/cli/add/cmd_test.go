package add

import (
	"path/filepath"
	"testing"
)

func TestResolveDestination(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPath string
	}{
		{name: "default destination", input: "", wantPath: DefaultDestination},
		{name: "current directory", input: ".", wantPath: "."},
		{name: "relative path", input: "./src/utils", wantPath: filepath.Clean("./src/utils")},
		{name: "parent directory", input: "..", wantPath: ".."},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveDestination(tt.input)
			if got != tt.wantPath {
				t.Fatalf("ResolveDestination(%q) = %q, want %q", tt.input, got, tt.wantPath)
			}
		})
	}
}

func TestResolveAddResourceType(t *testing.T) {
	tests := []struct {
		name        string
		stackFlag   bool
		snippetFlag bool
		want        ResourceType
		wantErr     bool
	}{
		{name: "auto by default", stackFlag: false, snippetFlag: false, want: ResourceTypeAuto, wantErr: false},
		{name: "force stack", stackFlag: true, snippetFlag: false, want: ResourceTypeStack, wantErr: false},
		{name: "force snippet", stackFlag: false, snippetFlag: true, want: ResourceTypeSnippet, wantErr: false},
		{name: "mutually exclusive flags", stackFlag: true, snippetFlag: true, want: ResourceTypeAuto, wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opts := Options{AsStack: tt.stackFlag, AsSnippet: tt.snippetFlag}
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
