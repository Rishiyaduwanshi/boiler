package utils

import "testing"

func TestIsCommandVarReference(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "valid variable reference", input: "bl__TEAM_REG", want: true},
		{name: "hyphen allowed", input: "bl__team-reg", want: true},
		{name: "email is not var", input: "alice@example.com", want: false},
		{name: "version token is not var", input: "express@1", want: false},
		{name: "legacy @ prefix is not var", input: "@TEAM_REG", want: false},
		{name: "invalid chars", input: "bl__team.reg", want: true}, // 'bl__team' is the valid var part
		{name: "embedded valid var", input: "github/repo/bl__file", want: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCommandVarReference(tt.input); got != tt.want {
				t.Fatalf("IsCommandVarReference(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeVarKey(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain key", input: "API_URL", want: "api_url"},
		{name: "snippet prefixed", input: "bl__API_URL", want: "api_url"},
		{name: "hyphen preserved", input: "team-reg", want: "team-reg"},
		{name: "empty key", input: "", wantErr: true},
		{name: "invalid key", input: "api.url", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeVarKey(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeVarKey(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeVarKey(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeVarKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLookupVar(t *testing.T) {
	vars := map[string]string{
		"api_url":  "https://api.example.com",
		"TEAM_REG": "https://github.com/myorg/boiler",
		"team-reg": "https://registry.example.com",
	}

	value, ok, err := LookupVar(vars, "bl__API_URL")
	if err != nil {
		t.Fatalf("LookupVar returned error: %v", err)
	}
	if !ok {
		t.Fatal("LookupVar should find bl__API_URL")
	}
	if value != "https://api.example.com" {
		t.Fatalf("LookupVar value = %q", value)
	}

	value, ok, err = LookupVar(vars, ":team_reg")
	if err != nil {
		t.Fatalf("LookupVar returned error: %v", err)
	}
	if !ok {
		t.Fatal("LookupVar should find :team_reg")
	}
	if value != "https://github.com/myorg/boiler" {
		t.Fatalf("LookupVar value = %q", value)
	}

	value, ok, err = LookupVar(vars, ":TEAM-reg")
	if err != nil {
		t.Fatalf("LookupVar returned error: %v", err)
	}
	if !ok {
		t.Fatal("LookupVar should find :TEAM-reg")
	}
	if value != "https://registry.example.com" {
		t.Fatalf("LookupVar value = %q", value)
	}
}

func TestResolveInlineVars(t *testing.T) {
	vars := map[string]string{"team_reg": "https://github.com/myorg/boiler", "lang": "Node"}

	resolved, resolvedFromVar, err := ResolveInlineVars("bl__team_reg", vars)
	if err != nil {
		t.Fatalf("ResolveInlineVars returned error: %v", err)
	}
	if !resolvedFromVar {
		t.Fatal("expected resolvedFromVar=true")
	}
	if resolved != "https://github.com/myorg/boiler" {
		t.Fatalf("resolved = %q", resolved)
	}

	resolved, resolvedFromVar, err = ResolveInlineVars("alice@example.com", vars)
	if err != nil {
		t.Fatalf("ResolveInlineVars returned error: %v", err)
	}
	if resolvedFromVar {
		t.Fatal("email token should not be treated as a variable reference")
	}
	if resolved != "alice@example.com" {
		t.Fatalf("resolved = %q", resolved)
	}

	// Test interpolation
	resolved, resolvedFromVar, err = ResolveInlineVars("github/gitignore:bl__lang.gitignore", vars)
	if err != nil {
		t.Fatalf("ResolveInlineVars returned error: %v", err)
	}
	if !resolvedFromVar {
		t.Fatal("expected resolvedFromVar=true")
	}
	if resolved != "github/gitignore:Node.gitignore" {
		t.Fatalf("resolved = %q", resolved)
	}
}

func TestResolveSnippetVarDefault(t *testing.T) {
	vars := map[string]string{"api_url": "https://api.example.com"}

	if got := ResolveSnippetVarDefault("bl__API_URL", "http://localhost", vars); got != "https://api.example.com" {
		t.Fatalf("ResolveSnippetVarDefault got %q", got)
	}
	if got := ResolveSnippetVarDefault("bl__MISSING", "fallback", vars); got != "fallback" {
		t.Fatalf("ResolveSnippetVarDefault fallback got %q", got)
	}
}
