package utils

import (
	"sort"
	"testing"
)

// ── DefaultIgnorePatterns ─────────────────────────────────────────────────────

func TestDefaultIgnorePatterns(t *testing.T) {
	expected := []string{"node_modules", ".git", ".DS_Store", "Thumbs.db"}
	if len(DefaultIgnorePatterns) != len(expected) {
		t.Fatalf("len(DefaultIgnorePatterns) = %d, want %d", len(DefaultIgnorePatterns), len(expected))
	}
	for _, name := range expected {
		found := false
		for _, p := range DefaultIgnorePatterns {
			if p == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DefaultIgnorePatterns missing %q", name)
		}
	}
}

// ── ParseResourceName (wrapper) ───────────────────────────────────────────────

func TestParseResourceNameWrapper(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"logger@1.js", "logger@1.js"},
		{"logger.js", "logger.js"}, // no version → name+ext
		{"express@2", "express@2"},
		{"express", "express"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseResourceName(tt.input)
			if got != tt.want {
				t.Errorf("ParseResourceName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ── FindMatchingResources ─────────────────────────────────────────────────────

func TestFindMatchingResources(t *testing.T) {
	all := []string{
		"logger@1.js",
		"logger@2.js",
		"logger@3.ts",
		"errorHandler@1.js",
		"express@1",
		"express@2",
	}

	t.Run("match by name and ext", func(t *testing.T) {
		got := FindMatchingResources(all, "logger", ".js")
		want := []string{"logger@1.js", "logger@2.js"}
		sortStrings(got)
		sortStrings(want)
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("match by name only (any ext)", func(t *testing.T) {
		got := FindMatchingResources(all, "logger", "")
		want := []string{"logger@1.js", "logger@2.js", "logger@3.ts"}
		sortStrings(got)
		sortStrings(want)
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("match stacks (no ext)", func(t *testing.T) {
		got := FindMatchingResources(all, "express", "")
		want := []string{"express@1", "express@2"}
		sortStrings(got)
		sortStrings(want)
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("no match returns nil", func(t *testing.T) {
		got := FindMatchingResources(all, "nonexistent", ".js")
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("empty source list", func(t *testing.T) {
		got := FindMatchingResources([]string{}, "logger", ".js")
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})
}

// ── PickFromList ──────────────────────────────────────────────────────────────

func TestPickFromList_SingleItem(t *testing.T) {
	// Single item → auto-selects without any stdin interaction
	got, err := PickFromList("logger.js", []string{"logger@1.js"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "logger@1.js" {
		t.Errorf("got %q, want %q", got, "logger@1.js")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func sortStrings(s []string) {
	sort.Strings(s)
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestIsValidAliasKey(t *testing.T) {
	valid := []string{"ll", "gi", "my_alias", "my-alias", "a1"}
	for _, name := range valid {
		if !IsValidAliasKey(name) {
			t.Errorf("expected %q to be valid alias key", name)
		}
	}

	invalid := []string{"LL", "MyAlias", "1alias", "_alias", "alias space", ""}
	for _, name := range invalid {
		if IsValidAliasKey(name) {
			t.Errorf("expected %q to be invalid alias key", name)
		}
	}
}

func TestIsValidVarKey(t *testing.T) {
	valid := []string{"api_url", "bl__api_url", "org", "repo_1"}
	for _, key := range valid {
		if !IsValidVarKey(key) {
			t.Errorf("expected %q to be valid var key", key)
		}
	}

	invalid := []string{"API_URL", "bl__API_URL", "var-dash", ""}
	for _, key := range invalid {
		if IsValidVarKey(key) {
			t.Errorf("expected %q to be invalid var key", key)
		}
	}
}

func TestIsValidScope(t *testing.T) {
	if !IsValidScope("global") || !IsValidScope("local") || !IsValidScope("") {
		t.Error("global, local, and empty should be valid scopes")
	}
	if IsValidScope("invalid_scope") {
		t.Error("invalid_scope should be invalid")
	}
}
