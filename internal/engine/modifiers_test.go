package engine

import (
	"testing"
)

func TestApplyModifiers(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		modifiers []string
		expected  string
	}{
		{
			name:      "No modifiers",
			value:     "hello world",
			modifiers: []string{},
			expected:  "hello world",
		},
		{
			name:      "Lowercase",
			value:     "Hello World",
			modifiers: []string{"lowercase"},
			expected:  "hello world",
		},
		{
			name:      "Uppercase",
			value:     "hello world",
			modifiers: []string{"uppercase"},
			expected:  "HELLO WORLD",
		},
		{
			name:      "Capitalize",
			value:     "hello world",
			modifiers: []string{"capitalize"},
			expected:  "Hello world",
		},
		{
			name:      "Capitalize with alias",
			value:     "hello world",
			modifiers: []string{"capitalise"},
			expected:  "Hello world",
		},
		{
			name:      "Snake Case",
			value:     "HelloWorld",
			modifiers: []string{"snake_case"},
			expected:  "hello_world",
		},
		{
			name:      "Snake Case with alias",
			value:     "HelloWorld",
			modifiers: []string{"snakeCase"},
			expected:  "hello_world",
		},
		{
			name:      "Camel Case",
			value:     "hello-world",
			modifiers: []string{"camel_case"},
			expected:  "helloWorld",
		},
		{
			name:      "Kebab Case",
			value:     "HelloWorld",
			modifiers: []string{"kebab_case"},
			expected:  "hello-world",
		},
		{
			name:      "Multiple Modifiers (camelCase -> pascalCase)",
			value:     "my-awesome-route",
			modifiers: []string{"camelCase", "pascalCase"},
			expected:  "MyAwesomeRoute",
		},
		{
			name:      "Multiple Modifiers (snake_case -> uppercase)",
			value:     "myAwesomeRoute",
			modifiers: []string{"snake_case", "uppercase"},
			expected:  "MY_AWESOME_ROUTE",
		},
		{
			name:      "Unknown Modifier (should be ignored)",
			value:     "hello",
			modifiers: []string{"unknown_format", "capitalize"},
			expected:  "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyModifiers(tt.value, tt.modifiers)
			if result != tt.expected {
				t.Errorf("ApplyModifiers(%q, %v) = %q; expected %q", tt.value, tt.modifiers, result, tt.expected)
			}
		})
	}
}
