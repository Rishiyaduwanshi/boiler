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
		case "lowercase", "toLowerCase", "to_lower_case":
			value = utils.Lowercase(value)
		case "uppercase", "toUpperCase", "to_upper_case":
			value = utils.Uppercase(value)
		case "title", "titleCase", "title_case":
			value = utils.Title(value)
		case "camel_case", "camelCase":
			value = utils.CamelCase(value)
		case "pascal_case", "pascalCase", "PascalCase":
			value = utils.PascalCase(value)
		case "snake_case", "snakeCase":
			value = utils.SnakeCase(value)
		case "kebab_case", "kebabCase":
			value = utils.KebabCase(value)
			// We can easily add has_flag, is_empty here if they return strings,
			// but boolean helpers are usually evaluated in conditions, not string modifiers.
		}
	}
	return value
}
