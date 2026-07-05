package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// CommonMetadata represents shared metadata between stack and snippet
type CommonMetadata struct {
	Name        string
	Version     string
	Author      string
	Description string
	CreatedAt   time.Time
}

// SnippetMetadata includes snippet-specific fields
type SnippetMetadata struct {
	CommonMetadata
	Language  string
	Variables map[string]string // bl__VAR_NAME -> default value
}

// PromptCommonMetadata prompts user for common metadata fields
func PromptCommonMetadata(defaultName string, skipPrompts bool) (CommonMetadata, error) {
	var meta CommonMetadata

	// Get author from git
	author := getGitAuthor()

	if skipPrompts {
		meta = CommonMetadata{
			Name:        defaultName,
			Version:     "1",
			Author:      author,
			Description: "",
			CreatedAt:   time.Now(),
		}
	} else {
		// Use defaultName directly, don't prompt
		meta.Name = defaultName
		meta.Description = PromptString("Description (optional)", "")
		meta.Author = PromptString("Author", author)
		meta.Version = PromptString("Version", "1")
		meta.CreatedAt = time.Now()
	}

	return meta, nil
}

// getGitAuthor fetches git user name
func getGitAuthor() string {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// ParseSnippetMetadata reads a snippet file and extracts metadata from comments
func ParseSnippetMetadata(filePath string) (*SnippetMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	meta := &SnippetMetadata{
		Variables: make(map[string]string),
	}

	scanner := bufio.NewScanner(file)

	// Regex patterns for metadata
	authorRe := regexp.MustCompile(`__author\s+(.+)`)
	descRe := regexp.MustCompile(`__desc\s+(.+)`)
	versionRe := regexp.MustCompile(`__version\s+(.+)`)
	// Match variable names with underscores (e.g., bl__API_URL)
	varRe := regexp.MustCompile(`__var\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*(.+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Extract metadata (regex handles any comment style)
		if matches := authorRe.FindStringSubmatch(line); len(matches) > 1 {
			meta.Author = strings.TrimSpace(matches[1])
		}
		if matches := descRe.FindStringSubmatch(line); len(matches) > 1 {
			meta.Description = strings.TrimSpace(matches[1])
		}
		if matches := versionRe.FindStringSubmatch(line); len(matches) > 1 {
			meta.Version = strings.TrimSpace(matches[1])
		}
		if matches := varRe.FindStringSubmatch(line); len(matches) > 2 {
			varName := strings.TrimSpace(matches[1])
			varValue := strings.TrimSpace(matches[2])
			meta.Variables[varName] = varValue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return meta, nil
}

// ValidateSnippetMetadata fills in optional fields where possible and warns
// when attribution data is missing. It does not block storing a snippet.
func ValidateSnippetMetadata(meta *SnippetMetadata) error {
	if meta.Author == "" {
		meta.Author = getGitAuthor()
	}
	if meta.Author == "" {
		fmt.Fprintln(os.Stderr, "warning: no __author found; add '// __author Your Name' for attribution")
	}
	// Name, Description, and Version are optional.
	// Version auto-increments based on existing versions in store.
	return nil
}

// GenerateSnippetTemplate creates a snippet file with metadata comments
func GenerateSnippetTemplate(filePath string, meta CommonMetadata, prefix, suffix string) error {
	sfx := ""
	if suffix != "" {
		sfx = " " + suffix
	}
	content := fmt.Sprintf("%s  __author %s%s\n%s  __desc %s%s\n%s  __var bl__EXAMPLE_VAR = DefaultValue%s\n\n%s  Your code here%s\n",
		prefix, meta.Author, sfx,
		prefix, meta.Description, sfx,
		prefix, sfx,
		prefix, sfx,
	)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
