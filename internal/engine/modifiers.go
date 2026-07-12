package engine

import (
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

// ApplyModifiers takes a base string value and applies a list of modifiers in order.
// Modifiers are applied sequentially.
// Example: value="hello world", modifiers=["camelCase", "capitalize"]
func ApplyModifiers(value string, modifiers []string) string {
	if len(modifiers) == 0 {
		return value
	}

	for _, mod := range modifiers {
		mod = strings.TrimSpace(mod)
		switch mod {
		case "capitalize", "capitalise":
			value = utils.Capitalize(value)
		case "lower", "lowerCase", "lower_case":
			value = utils.Lowercase(value)
		case "upper", "upperCase", "upper_case":
			value = utils.Uppercase(value)
		case "title", "titleCase", "title_case":
			value = utils.Title(value)
		case "camelCase", "camel_case":
			value = utils.CamelCase(value)
		case "pascalCase", "pascal_case":
			value = utils.PascalCase(value)
		case "snakeCase", "snake_case":
			value = utils.SnakeCase(value)
		case "kebabCase", "kebab_case":
			value = utils.KebabCase(value)
			// We can easily add has_flag, is_empty here if they return strings,
			// but boolean helpers are usually evaluated in conditions, not string modifiers.
		}
	}
	return value
}

// ResolveVariable takes an expression like "bl__1.capitalize().snake_case()"
// and a map of variables, looks up the base variable, and applies all modifiers.
func ResolveVariable(expr string, vars map[string]string) string {
	expr = strings.TrimSpace(expr)
	parts := strings.Split(expr, ".")

	baseVarName := parts[0]
	val := vars[baseVarName]

	if len(parts) > 1 {
		// Remove empty parenthesis from modifiers if present (e.g., .capitalize() -> capitalize)
		modifiers := make([]string, len(parts)-1)
		for i, mod := range parts[1:] {
			modifiers[i] = strings.TrimSuffix(mod, "()")
		}
		val = ApplyModifiers(val, modifiers)
	}
	return val
}
