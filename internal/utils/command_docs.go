package utils

import (
	"strings"

	"github.com/spf13/cobra"
)

// NormalizeCommandTreeDocs normalizes help/documentation text for all commands.
func NormalizeCommandTreeDocs(cmd *cobra.Command) {
	cmd.Short = normalizeCommandShortText(cmd.Short)
	cmd.Long = normalizeCommandDocBlock(cmd.Long)
	cmd.Example = normalizeCommandDocBlock(cmd.Example)

	for _, child := range cmd.Commands() {
		NormalizeCommandTreeDocs(child)
	}
}

func normalizeCommandShortText(text string) string {
	normalized := strings.ReplaceAll(text, "\t", " ")
	return strings.TrimSpace(normalized)
}

func normalizeCommandDocBlock(block string) string {
	if block == "" {
		return ""
	}

	normalized := strings.ReplaceAll(block, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	for i, line := range lines {
		line = strings.ReplaceAll(line, "\t", "  ")
		line = strings.TrimRight(line, " ")

		if strings.TrimSpace(line) == "" {
			lines[i] = ""
			continue
		}

		leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
		if leadingSpaces > 2 {
			line = "  " + strings.TrimLeft(line, " ")
		}

		lines[i] = line
	}

	return strings.Trim(strings.Join(lines, "\n"), "\n")
}
