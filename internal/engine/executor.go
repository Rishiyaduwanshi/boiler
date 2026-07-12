package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
		// Raw (non-interpolated) run line: tokenize first, then interpolate per-token
		cmdStr := strings.TrimSpace(strings.TrimPrefix(commandLine, "run "))
		fmt.Printf("[ENGINE] Running OS command: %s\n", cmdStr)
		return executeOSCommand(state, cmdStr)
	} else {
		// Assume it's a native Boiler command (use, add, var, alias, etc.)
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to resolve boiler executable: %w", err)
		}

		fmt.Printf("[ENGINE] Running Boiler command: %s %s\n", filepath.Base(exePath), commandLine)

		// Tokenize the template FIRST, then interpolate each token individually.
		// This keeps variable values (e.g. paths with spaces) as atomic arguments.
		templateArgs := parseArgs(commandLine)
		args := make([]string, len(templateArgs))
		for i, arg := range templateArgs {
			if state != nil {
				args[i] = state.InterpolateVariables(arg)
			} else {
				args[i] = arg
			}
		}
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

// parseArgs safely splits a command string by spaces, respecting both single
// and double quotes, as well as backslash-escaped characters.
func parseArgs(s string) []string {
	var args []string
	var current []rune
	var quoteChar rune
	inQuote := false
	escaped := false

	for _, r := range s {
		if escaped {
			current = append(current, r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if inQuote {
			if r == quoteChar {
				inQuote = false
			} else {
				current = append(current, r)
			}
		} else {
			switch r {
			case '"', '\'':
				inQuote = true
				quoteChar = r
			case ' ', '\t':
				if len(current) > 0 {
					args = append(args, string(current))
					current = current[:0]
				}
			default:
				current = append(current, r)
			}
		}
	}
	if len(current) > 0 {
		args = append(args, string(current))
	}
	return args
}

// executeOSCommand invokes a real executable directly without involving any
// shell interpreter. This prevents command injection on all platforms:
// user-supplied template variables can never be interpreted as shell
// metacharacters because no shell is involved in the execution.
//
// Variable values are interpolated PER TOKEN (after tokenization), so a value
// like "left-pad --save-dev" remains a single argv entry and cannot inject flags.
//
// Note: shell built-ins (e.g. echo, dir, cd) are intentionally not supported.
// .bl scripts should call real executables available in PATH (npm, git, go, etc.).
func executeOSCommand(state *ScriptState, cmdStr string) error {
	// Tokenize the template BEFORE variable substitution
	templateArgs := parseArgs(cmdStr)
	if len(templateArgs) == 0 {
		return fmt.Errorf("run: empty command")
	}

	// Interpolate each token individually — variable values stay atomic
	args := make([]string, len(templateArgs))
	for i, arg := range templateArgs {
		if state != nil {
			args[i] = state.InterpolateVariables(arg)
		} else {
			args[i] = arg
		}
	}

	exe, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("run: executable %q not found in PATH: %w", args[0], err)
	}
	cmd := exec.Command(exe, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
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
