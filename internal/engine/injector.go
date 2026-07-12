package engine

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// InjectCode parses an inject command and modifies the target file safely using markers.
func InjectCode(cmd string) error {
	// 1. Extract content from backticks
	parts := strings.SplitN(cmd, "`", 3)
	if len(parts) != 3 {
		return fmt.Errorf("invalid inject command: missing or unclosed backticks for content")
	}
	content := parts[1]

	// 2. Parse the rest of the command before the backticks
	cmdWithoutContent := strings.TrimSpace(parts[0])

	tokens := strings.Fields(cmdWithoutContent)
	if len(tokens) < 2 {
		return fmt.Errorf("invalid inject command: missing target file")
	}

	filePath := tokens[1]
	detector := ""
	appendMode := true // Default to append mode

	for i := 2; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == "-d" || tok == "--detect" {
			if i+1 < len(tokens) {
				detector = tokens[i+1]
				i++
			}
		} else if tok == "-a" || tok == "--append" {
			appendMode = true
		} else if tok == "-p" || tok == "--prepend" {
			appendMode = false
		} else if tok == "-c" || tok == "--content" {
			// Content is parsed via backticks, so ignore the flag itself
		}
	}

	if detector == "" {
		return fmt.Errorf("missing detector name (-d or --detect) in inject command")
	}

	return PerformInjection(filePath, detector, content, appendMode)
}

// PerformInjection finds the START/END markers in a file and injects the content.
func PerformInjection(filePath, detector, content string, appendMode bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open target file for injection: %w", err)
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read target file: %w", err)
	}

	startMarker := fmt.Sprintf("bl__DETECTOR_START_%s", detector)
	endMarker := fmt.Sprintf("bl__DETECTOR_END_%s", detector)

	var newLines []string
	injected := false
	contentLines := strings.Split(strings.Trim(content, "\r\n"), "\n")

	for _, line := range lines {
		// If appending, inject the content JUST BEFORE the END marker
		if !injected && appendMode && strings.Contains(line, endMarker) {
			newLines = append(newLines, contentLines...)
			injected = true
		}

		newLines = append(newLines, line)

		// If prepending, inject the content JUST AFTER the START marker
		if !injected && !appendMode && strings.Contains(line, startMarker) {
			newLines = append(newLines, contentLines...)
			injected = true
		}
	}

	if !injected {
		return fmt.Errorf("detector markers for '%s' not found in %s", detector, filePath)
	}

	out := strings.Join(newLines, "\n") + "\n"
	return os.WriteFile(filePath, []byte(out), 0644)
}
