package utils

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	camelCaseRe = regexp.MustCompile(`(?:^|[-_])(\w)`)
	snakeCaseRe = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

// Capitalize makes the first letter uppercase and the rest lowercase.
func Capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	runes := []rune(s)
	return string(unicode.ToUpper(runes[0])) + strings.ToLower(string(runes[1:]))
}

// Lowercase converts the entire string to lowercase.
func Lowercase(s string) string {
	return strings.ToLower(s)
}

// Uppercase converts the entire string to uppercase.
func Uppercase(s string) string {
	return strings.ToUpper(s)
}

// Title converts the string to Title Case (Every Word Capitalized).
func Title(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			words[i] = string(unicode.ToUpper(runes[0])) + string(runes[1:])
		}
	}
	return strings.Join(words, " ")
}

// CamelCase converts snake_case, kebab-case, or space separated words to camelCase.
func CamelCase(s string) string {
	if s == "" {
		return ""
	}

	// Convert all special characters and spaces to underscores first
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	// Apply regex to convert _x to X
	camel := camelCaseRe.ReplaceAllStringFunc(s, func(match string) string {
		return strings.ToUpper(strings.TrimLeft(match, "-_"))
	})

	// Ensure the very first character is lowercase
	if len(camel) > 0 {
		runes := []rune(camel)
		return string(unicode.ToLower(runes[0])) + string(runes[1:])
	}
	return camel
}

// PascalCase (UpperCamelCase) is like camelCase but the first letter is capitalized.
func PascalCase(s string) string {
	camel := CamelCase(s)
	if len(camel) > 0 {
		runes := []rune(camel)
		return string(unicode.ToUpper(runes[0])) + string(runes[1:])
	}
	return camel
}

// SnakeCase converts CamelCase, kebab-case, or space separated words to snake_case.
func SnakeCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// Insert underscores between lower/number and Upper case
	s = snakeCaseRe.ReplaceAllString(s, "${1}_${2}")

	// Convert spaces and hyphens to underscores
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	return strings.ToLower(s)
}

// KebabCase converts CamelCase, snake_case, or space separated words to kebab-case.
func KebabCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// Insert hyphens between lower/number and Upper case
	s = snakeCaseRe.ReplaceAllString(s, "${1}-${2}")

	// Convert spaces and underscores to hyphens
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")

	return strings.ToLower(s)
}
