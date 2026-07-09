package engine

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ScriptState holds the execution context of a .bl script.
type ScriptState struct {
	Vars      map[string]string
	SkipStack []bool // Keeps track of `#if` and `#else` skipping logic
}

// isSkipping returns true if any block in the SkipStack is currently false.
func (s *ScriptState) isSkipping() bool {
	for _, shouldExecute := range s.SkipStack {
		if !shouldExecute {
			return true
		}
	}
	return false
}

// varInterpolationRe matches variable references like bl__foo, bl__foo.capitalize(), etc.
// It explicitly looks for () at the end of modifiers so it doesn't match file extensions like .js
var varInterpolationRe = regexp.MustCompile(`(bl__[a-zA-Z0-9_]+(?:\.[a-zA-Z0-9_]+\(\))*)`)

// InterpolateVariables replaces bl__ variables in a string with their actual values.
func (s *ScriptState) InterpolateVariables(line string) string {
	return varInterpolationRe.ReplaceAllStringFunc(line, func(match string) string {
		return ResolveVariable(match, s.Vars)
	})
}

// ParseAndExecute reads a .bl file and executes it line-by-line.
func ParseAndExecute(filePath string, initialVars map[string]string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open script %s: %w", filePath, err)
	}
	defer file.Close()

	state := &ScriptState{
		Vars:      make(map[string]string),
		SkipStack: make([]bool, 0),
	}

	for k, v := range initialVars {
		state.Vars[k] = v
	}

	scanner := bufio.NewScanner(file)

	inBacktick := false
	var backtickContent strings.Builder
	var backtickCommand string

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		rawLine := scanner.Text()

		if inBacktick {
			if strings.Contains(rawLine, "`") {
				parts := strings.SplitN(rawLine, "`", 2)
				backtickContent.WriteString(parts[0])
				inBacktick = false

				fullCommand := backtickCommand + "`" + backtickContent.String() + "`" + parts[1]
				if err := executeLine(state, fullCommand, lineNumber); err != nil {
					return err
				}
				backtickContent.Reset()
			} else {
				backtickContent.WriteString(rawLine + "\n")
			}
			continue
		}

		line := strings.TrimSpace(rawLine)

		// Ignore empty lines and pure comments (not starting with #if, #else, #endif)
		if line == "" || (strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "#if") && !strings.HasPrefix(line, "#else") && !strings.HasPrefix(line, "#endif")) {
			continue
		}

		// Handle control flow (evaluated even when skipping to maintain stack)
		if strings.HasPrefix(line, "#if ") {
			cond := strings.TrimSpace(strings.TrimPrefix(line, "#if "))
			shouldExecute := EvaluateCondition(cond, state.Vars)

			if state.isSkipping() {
				shouldExecute = false
			}
			state.SkipStack = append(state.SkipStack, shouldExecute)
			continue
		} else if strings.HasPrefix(line, "#else") {
			if len(state.SkipStack) == 0 {
				return fmt.Errorf("error at line %d: #else without #if", lineNumber)
			}

			parentSkipping := false
			if len(state.SkipStack) > 1 {
				for _, p := range state.SkipStack[:len(state.SkipStack)-1] {
					if !p {
						parentSkipping = true
						break
					}
				}
			}

			curr := state.SkipStack[len(state.SkipStack)-1]
			if parentSkipping {
				state.SkipStack[len(state.SkipStack)-1] = false
			} else {
				state.SkipStack[len(state.SkipStack)-1] = !curr
			}
			continue
		} else if strings.HasPrefix(line, "#endif") {
			if len(state.SkipStack) == 0 {
				return fmt.Errorf("error at line %d: #endif without #if", lineNumber)
			}
			state.SkipStack = state.SkipStack[:len(state.SkipStack)-1]
			continue
		}

		if state.isSkipping() {
			continue
		}

		// Check if line starts a backtick block
		if strings.Contains(line, "`") {
			parts := strings.SplitN(line, "`", 2)
			backtickCommand = parts[0]

			if strings.Contains(parts[1], "`") {
				// Backticks open and close on the same line, just execute normally
			} else {
				inBacktick = true
				backtickContent.WriteString(parts[1] + "\n")
				continue
			}
		}

		if err := executeLine(state, line, lineNumber); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read script: %w", err)
	}

	if len(state.SkipStack) > 0 {
		return fmt.Errorf("unclosed #if block at end of file")
	}
	if inBacktick {
		return fmt.Errorf("unclosed backtick block at end of file")
	}

	return nil
}

func executeLine(state *ScriptState, line string, lineNumber int) error {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	// Handle variable assignment FIRST, before general interpolation
	if strings.HasPrefix(line, "__var ") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			varName := strings.TrimSpace(strings.TrimPrefix(parts[0], "__var "))

			// Interpolate the RIGHT side only
			varValue := state.InterpolateVariables(strings.TrimSpace(parts[1]))
			varValue = strings.Trim(varValue, `"'`)

			state.Vars[varName] = varValue
		}
		return nil
	}

	// Interpolate variables for all other executable lines
	line = state.InterpolateVariables(line)

	// Handle metadata (ignore for execution, just store if needed)
	if strings.HasPrefix(line, "__") {
		return nil
	}

	// Execute command
	if err := ExecuteCommand(state, line); err != nil {
		return fmt.Errorf("error at line %d: %w", lineNumber, err)
	}
	return nil
}
