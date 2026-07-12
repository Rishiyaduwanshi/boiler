package engine

import (
	"strings"
)

// EvaluateCondition parses and evaluates a logical condition string without AST.
// Supports: ==, !=, !, &&, || (Evaluated left-to-right, AND before OR, no brackets)
func EvaluateCondition(cond string, vars map[string]string) bool {
	if strings.TrimSpace(cond) == "" {
		return false
	}

	// First split by OR (||)
	orParts := strings.Split(cond, "||")
	for _, orPart := range orParts {
		// Inside each OR part, all AND (&&) parts must be true
		andParts := strings.Split(orPart, "&&")
		allAndsTrue := true

		for _, andPart := range andParts {
			andPart = strings.TrimSpace(andPart)
			if andPart == "" {
				continue
			}
			if !evaluateSingle(andPart, vars) {
				allAndsTrue = false
				break
			}
		}

		// If all AND parts in this OR block were true, the whole expression is true
		if allAndsTrue {
			return true
		}
	}

	return false
}

func evaluateSingle(cond string, vars map[string]string) bool {
	cond = strings.TrimSpace(cond)

	// Handle NOT operator
	isNot := false
	if strings.HasPrefix(cond, "!") {
		isNot = true
		cond = strings.TrimSpace(strings.TrimPrefix(cond, "!"))
	}

	result := false

	// Handle Equality
	if strings.Contains(cond, "==") {
		parts := strings.SplitN(cond, "==", 2)
		left := resolveVal(parts[0], vars)
		right := resolveVal(parts[1], vars)
		result = (left == right)
	} else if strings.Contains(cond, "!=") {
		parts := strings.SplitN(cond, "!=", 2)
		left := resolveVal(parts[0], vars)
		right := resolveVal(parts[1], vars)
		result = (left != right)
	} else {
		// Just a boolean variable check, e.g., "bl__is_ts"
		val := resolveVal(cond, vars)
		val = strings.ToLower(val)
		result = (val == "true" || val == "1" || val == "yes")
	}

	if isNot {
		return !result
	}
	return result
}

func resolveVal(s string, vars map[string]string) string {
	s = strings.TrimSpace(s)
	// Remove surrounding quotes if present
	s = strings.Trim(s, `"'`)

	// If it's a variable reference (starts with bl__), resolve it including modifiers
	if strings.HasPrefix(s, "bl__") {
		return ResolveVariable(s, vars)
	}
	return s
}
