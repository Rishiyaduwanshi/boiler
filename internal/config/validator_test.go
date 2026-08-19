package config

import "testing"

func TestConfigValidate_ValidConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Aliases["ll"] = "ls"
	cfg.Vars["api_url"] = "https://api.example.com"
	cfg.Scope = "global"

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestConfigValidate_InvalidScope(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Scope = "invalid_scope"

	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid scope, got nil")
	}
}

func TestConfigValidate_InvalidAliasKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Aliases["LL"] = "ls"

	if err := cfg.Validate(); err == nil {
		t.Error("expected error for uppercase alias key 'LL', got nil")
	}
}

func TestConfigValidate_EmptyAliasTarget(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Aliases["ll"] = "   "

	if err := cfg.Validate(); err == nil {
		t.Error("expected error for whitespace alias target, got nil")
	}
}

func TestConfigValidate_InvalidVarKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Vars["API_URL"] = "https://example.com"

	if err := cfg.Validate(); err == nil {
		t.Error("expected error for uppercase var key, got nil")
	}
}

func TestConfigValidate_EmptyPaths(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Paths = &Paths{Root: "  ", Store: "store"}

	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty paths.root, got nil")
	}
}

func TestConfigValidate_InvalidRegistryURL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Registry = "invalid-url-schema"

	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid registry URL, got nil")
	}
}

func TestConfigValidate_EmptyVarValueAllowed(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Vars["gic"] = ""

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected empty var value to be allowed, got error: %v", err)
	}
}

func TestConfigValidate_InvalidArtifactFormat(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Artifacts["go"] = "//" // Missing two-space separator

	if err := cfg.Validate(); err == nil {
		t.Error("expected error for artifact format without two-space separator, got nil")
	}
}
