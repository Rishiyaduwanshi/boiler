package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/constants"
)

const (
	CommandVarPrefix = ":"
)

var (
	cliVarReplaceRe    = regexp.MustCompile(fmt.Sprintf(`(?i)%s([a-z_][a-z0-9_-]*)`, constants.VarPrefix))
	normalizedVarKeyRe = regexp.MustCompile(constants.VarKeyPattern)
)

// IsCommandVarReference reports whether token contains a command variable reference (e.g. bl__TEAM_REG).
func IsCommandVarReference(token string) bool {
	return cliVarReplaceRe.MatchString(token)
}

// NormalizeVarKey converts a user-provided variable key to canonical form.
// Rules:
// - trim spaces
// - optional prefixes are stripped (: and bl__)
// - hyphens and underscores are preserved
// - key is lower-cased
func NormalizeVarKey(raw string) (string, error) {
	key := strings.TrimSpace(raw)
	if key == "" {
		return "", fmt.Errorf("variable name cannot be empty")
	}

	if strings.HasPrefix(key, CommandVarPrefix) {
		key = strings.TrimPrefix(key, CommandVarPrefix)
	}
	if hasPrefixFold(key, constants.VarPrefix) {
		key = key[len(constants.VarPrefix):]
	}

	key = strings.TrimSpace(key)
	key = strings.ToLower(key)

	if !normalizedVarKeyRe.MatchString(key) {
		return "", fmt.Errorf("invalid variable name %q: use letters, numbers, underscores, and hyphens", raw)
	}

	return key, nil
}

// NormalizeVarMap returns a new map with canonicalized variable keys.
func NormalizeVarMap(vars map[string]string) map[string]string {
	normalized := make(map[string]string)
	for k, v := range vars {
		canonicalKey, err := NormalizeVarKey(k)
		if err != nil {
			continue
		}
		normalized[canonicalKey] = v
	}
	return normalized
}

// LookupVar resolves a key using canonical variable-name matching.
func LookupVar(vars map[string]string, rawKey string) (string, bool, error) {
	key, err := NormalizeVarKey(rawKey)
	if err != nil {
		return "", false, err
	}
	if len(vars) == 0 {
		return "", false, nil
	}

	if value, ok := vars[key]; ok {
		return value, true, nil
	}

	// Fallback for any non-normalized in-memory maps.
	for existingKey, existingValue := range vars {
		normalizedExistingKey, normalizeErr := NormalizeVarKey(existingKey)
		if normalizeErr == nil && normalizedExistingKey == key {
			return existingValue, true, nil
		}
	}

	return "", false, nil
}

// ResolveInlineVars resolves bl__VAR references to their configured values inline.
// It replaces all occurrences of bl__VAR in the string.
func ResolveInlineVars(token string, vars map[string]string) (resolved string, resolvedFromVar bool, err error) {
	trimmedToken := strings.TrimSpace(token)
	if !IsCommandVarReference(trimmedToken) {
		return token, false, nil
	}

	var resolveErr error
	resolvedString := cliVarReplaceRe.ReplaceAllStringFunc(trimmedToken, func(match string) string {
		// match is e.g. bl__my_var. We pass it to LookupVar (which normalizes it and strips prefix).
		value, ok, err := LookupVar(vars, match)
		if err != nil {
			resolveErr = err
			return match
		}
		if !ok {
			resolveErr = fmt.Errorf("variable %q not found", match)
			return match
		}
		resolvedFromVar = true
		return value
	})

	if resolveErr != nil {
		return "", true, resolveErr
	}

	return resolvedString, resolvedFromVar, nil
}

// ResolveInputToken resolves :VAR references and returns field-specific errors.
func ResolveInputToken(token, field string, vars map[string]string) (string, error) {
	resolved, _, err := ResolveInlineVars(token, vars)
	if err != nil {
		return "", fmt.Errorf("invalid %s: %w", field, err)
	}
	return resolved, nil
}

// ResolveSnippetVarDefault returns config var value if present, otherwise fallback.
func ResolveSnippetVarDefault(varName, fallback string, vars map[string]string) string {
	value, ok, err := LookupVar(vars, varName)
	if err != nil || !ok {
		return fallback
	}
	return value
}

func hasPrefixFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return strings.EqualFold(s[:len(prefix)], prefix)
}
