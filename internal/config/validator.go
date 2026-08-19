package config

import (
	"fmt"
	"os"
	"sort"
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

// SanitizeAndWarn validates critical structural fields, prints warnings to os.Stderr
// for non-critical formatting issues (like mixed-case alias/var keys), and auto-normalizes
// them in memory so CLI startup is never blocked (allowing bl conf -e repair route).
func (c *Config) SanitizeAndWarn(filePath string) error {
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

	// Auto-normalize and warn for aliases with deterministic key resolution
	if len(c.Aliases) > 0 {
		keys := make([]string, 0, len(c.Aliases))
		for k := range c.Aliases {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		normalizedAliases := make(map[string]string, len(c.Aliases))
		for _, key := range keys {
			target := c.Aliases[key]
			if strings.TrimSpace(target) == "" {
				return fmt.Errorf("invalid alias target for %q: target cannot be empty", key)
			}
			lowerKey := strings.ToLower(strings.TrimSpace(key))
			if !utils.IsValidAliasKey(lowerKey) {
				return fmt.Errorf("invalid alias key %q: must use letters, numbers, underscores, and hyphens", key)
			}
			if key != lowerKey {
				fmt.Fprintf(os.Stderr, "Warning: invalid alias key %q in %s (must be lowercase). Auto-normalizing to %q. Run 'bl conf -e' to fix.\n", key, filePath, lowerKey)
			}
			// If exact lowercase key is already set, exact match takes precedence over mixed-case collision
			if _, exists := normalizedAliases[lowerKey]; !exists || key == lowerKey {
				normalizedAliases[lowerKey] = target
			}
		}
		c.Aliases = normalizedAliases
	}

	// Auto-normalize and warn for vars with deterministic key resolution
	if len(c.Vars) > 0 {
		keys := make([]string, 0, len(c.Vars))
		for k := range c.Vars {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		normalizedVars := make(map[string]string, len(c.Vars))
		for _, key := range keys {
			val := c.Vars[key]
			lowerKey := strings.ToLower(strings.TrimSpace(key))
			if !utils.IsValidVarKey(lowerKey) {
				return fmt.Errorf("invalid variable key %q: must be lowercase (a-z, 0-9, _)", key)
			}
			if key != lowerKey {
				fmt.Fprintf(os.Stderr, "Warning: invalid variable key %q in %s (must be lowercase). Auto-normalizing to %q. Run 'bl conf -e' to fix.\n", key, filePath, lowerKey)
			}
			if _, exists := normalizedVars[lowerKey]; !exists || key == lowerKey {
				normalizedVars[lowerKey] = val
			}
		}
		c.Vars = normalizedVars
	}

	return nil
}
