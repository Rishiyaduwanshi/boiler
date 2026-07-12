package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAndExecute(t *testing.T) {
	// We will create a temporary .bl file to parse and test
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "test.bl")

	dummyJsPath := filepath.Join(tempDir, "file.js")
	// Create dummy file for injection
	if err := os.WriteFile(dummyJsPath, []byte("// bl__DETECTOR_START_test\n// bl__DETECTOR_END_test\n"), 0644); err != nil {
		t.Fatalf("Failed to write dummy JS file: %v", err)
	}

	scriptContent := `
__desc = "Test Script"
__var bl__name = bl__1.capitalize()
__var bl__is_ts = true

#if bl__is_ts
	run go env GOPATH
	#if bl__name == "Abhinav"
		run go env GOOS
	#endif
#else
	run go env GOARCH
#endif

# Test backticks
inject ` + filepath.ToSlash(dummyJsPath) + ` -d test -a -c ` + "`" + `
console.log("bl__name");
` + "`" + `
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test script: %v", err)
	}

	initialVars := map[string]string{
		"bl__1": "abhinav", // will be capitalized to "Abhinav"
	}

	// ParseAndExecute should run without errors: variables resolved,
	// conditionals evaluated, inject applied, and OS commands executed.
	err = ParseAndExecute(scriptPath, initialVars)
	if err != nil {
		t.Fatalf("ParseAndExecute failed: %v", err)
	}
}

func TestInterpolateVariables(t *testing.T) {
	state := &ScriptState{
		Vars: map[string]string{
			"bl__foo": "hello world",
		},
	}

	tests := []struct {
		line     string
		expected string
	}{
		{"run echo bl__foo", "run echo hello world"},
		{"run echo bl__foo.capitalize()", "run echo Hello world"},
		{"run echo bl__foo.snake_case()", "run echo hello_world"},
		{"run echo bl__foo.snake_case().upperCase()", "run echo HELLO_WORLD"},
		{"use repo:file.js ./src/bl__foo.snake_case().js", "use repo:file.js ./src/hello_world.js"},
	}

	for _, tt := range tests {
		result := state.InterpolateVariables(tt.line)
		if result != tt.expected {
			t.Errorf("Interpolate(%q) = %q; expected %q", tt.line, result, tt.expected)
		}
	}
}
