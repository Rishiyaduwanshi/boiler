package config

import (
	"fmt"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

// Validate checks the entire Config struct against schema rules.
// Returns a detailed error if any field is invalid.
func (c *Config) Validate() error {
	if c == nil {
		return nil
	}

	if !utils.IsValidScope(c.Scope) {
		return fmt.Errorf("invalid scope %q: must be 'global' or 'local'", c.Scope)
	}

	if c.Registry != "" && !utils.IsURL(c.Registry) {
		return fmt.Errorf("invalid registry URL %q: must start with http:// or https://", c.Registry)
	}

	if c.DefaultEditor != "" && strings.TrimSpace(c.DefaultEditor) == "" {
		return fmt.Errorf("invalid defaultEditor: cannot be empty or whitespace")
	}

	if c.Paths != nil {
		if strings.TrimSpace(c.Paths.Root) == "" {
			return fmt.Errorf("invalid paths.root: cannot be empty")
		}
		if strings.TrimSpace(c.Paths.Store) == "" {
			return fmt.Errorf("invalid paths.store: cannot be empty")
		}
	}

	for key, val := range c.Artifacts {
		if strings.TrimSpace(key) == "" || strings.ToLower(key) != key {
			return fmt.Errorf("invalid artifact key %q: must be lowercase extension (e.g. go, py, md)", key)
		}
		if !strings.Contains(val, "  ") {
			return fmt.Errorf("invalid artifact format for %q: must contain two-space separator '  ' (e.g. '//  ' or '/*  */')", key)
		}
	}

	for key, target := range c.Aliases {
		if !utils.IsValidAliasKey(key) {
			return fmt.Errorf("invalid alias key %q: must be lowercase (a-z, 0-9, _, -)", key)
		}
		if strings.TrimSpace(target) == "" {
			return fmt.Errorf("invalid alias target for %q: target cannot be empty", key)
		}
	}

	for key := range c.Vars {
		if !utils.IsValidVarKey(key) {
			return fmt.Errorf("invalid variable key %q: must be lowercase (a-z, 0-9, _)", key)
		}
	}

	return nil
}
