package engine

import (
	"testing"
)

func TestEvaluateCondition(t *testing.T) {
	vars := map[string]string{
		"bl__is_ts": "true",
		"bl__is_js": "false",
		"bl__db":    "mongodb",
		"bl__port":  "8080",
		"bl__name":  "User",
	}

	tests := []struct {
		name     string
		cond     string
		expected bool
	}{
		{"Empty condition", "", false},
		{"Simple true boolean", "bl__is_ts", true},
		{"Simple false boolean", "bl__is_js", false},
		{"Equality check (string)", "bl__db == mongodb", true},
		{"Equality check with quotes", "bl__db == \"mongodb\"", true},
		{"Inequality check", "bl__port != 3000", true},
		{"NOT true", "!bl__is_ts", false},
		{"NOT false", "!bl__is_js", true},
		{"AND condition (true)", "bl__is_ts && bl__db == mongodb", true},
		{"AND condition (false)", "bl__is_ts && bl__db == postgres", false},
		{"OR condition (true)", "bl__is_js || bl__db == mongodb", true},
		{"OR condition (false)", "bl__is_js || bl__db == postgres", false},
		{"Complex AND/OR", "bl__is_js && bl__port == 8080 || bl__db == mongodb", true}, // (false AND true) OR true = true
		{"Variable with modifier", "bl__name.lowercase() == user", true},
		{"Variable with modifier equality", "bl__name.lowercase() == bl__name.lowercase()", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EvaluateCondition(tt.cond, vars)
			if result != tt.expected {
				t.Errorf("EvaluateCondition(%q) = %v; expected %v", tt.cond, result, tt.expected)
			}
		})
	}
}
