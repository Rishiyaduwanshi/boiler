package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ExecuteCommand routes parsed Boiler commands to their respective actions.
// Supports: use, run, inject
func ExecuteCommand(state *ScriptState, commandLine string) error {
	commandLine = strings.TrimSpace(commandLine)

	if strings.HasPrefix(commandLine, "inject ") {
		return InjectCode(commandLine)
	} else if strings.HasPrefix(commandLine, "create ") {
		return CreateFile(commandLine)
	} else if strings.HasPrefix(commandLine, "run ") {
		// Run an arbitrary OS command (e.g. npm install)
		cmdStr := strings.TrimPrefix(commandLine, "run ")
		cmdStr = strings.TrimSpace(cmdStr)
		fmt.Printf("[ENGINE] Running OS command: %s\n", cmdStr)
		return executeOSCommand(cmdStr)
	} else {
		// Assume it's a native Boiler command (use, add, var, alias, etc.)
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to resolve boiler executable: %w", err)
		}

		fmt.Printf("[ENGINE] Running Boiler command: %s %s\n", filepath.Base(exePath), commandLine)

		// Bypass OS shell entirely for native commands to avoid quoting issues
		args := parseArgs(commandLine)
		cmd := exec.Command(exePath, args...)

		// Attach standard streams so prompts and outputs work normally
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		// Pass ALL script variables to the child process via Environment Variables
		// We prefix them with BOILER_VAR_ so the child process's config loader can pick them up.
		cmd.Env = os.Environ()
		if state != nil && state.Vars != nil {
			for k, v := range state.Vars {
				envKey := fmt.Sprintf("BOILER_VAR_%s", k)
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", envKey, v))
			}
		}

		return cmd.Run()
	}
}

// parseArgs safely splits a command string by spaces, respecting double quotes.
func parseArgs(s string) []string {
	var args []string
	var current []rune
	inQuote := false
	for _, r := range s {
		if r == '"' {
			inQuote = !inQuote
		} else if r == ' ' && !inQuote {
			if len(current) > 0 {
				args = append(args, string(current))
				current = []rune{}
			}
		} else {
			current = append(current, r)
		}
	}
	if len(current) > 0 {
		args = append(args, string(current))
	}
	return args
}

func executeOSCommand(cmdStr string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CreateFile parses a create command and writes the multiline content to a new file.
func CreateFile(cmd string) error {
	// Extract content from backticks
	parts := strings.SplitN(cmd, "`", 3)
	if len(parts) != 3 {
		return fmt.Errorf("invalid create command: missing or unclosed backticks for content")
	}
	content := parts[1]

	// Parse the rest of the command
	cmdWithoutContent := strings.TrimSpace(parts[0])
	tokens := strings.Fields(cmdWithoutContent)
	if len(tokens) < 2 {
		return fmt.Errorf("invalid create command: missing target file")
	}

	filePath := tokens[1]

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", filePath, err)
	}

	return os.WriteFile(filePath, []byte(strings.TrimPrefix(content, "\n")), 0644)
}
