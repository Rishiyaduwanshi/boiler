package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/store"
)

// DefaultIgnorePatterns are common paths to skip when copying stacks.
var DefaultIgnorePatterns = []string{"node_modules", ".git", ".DS_Store", "Thumbs.db"}

// LoadStore creates and loads the store instance
func LoadStore(storePath string) (*store.Store, error) {
	st := store.NewStore(storePath)
	if err := st.Load(); err != nil {
		return nil, fmt.Errorf("failed to load store: %w", err)
	}
	return st, nil
}

// ConfirmAction prompts user for yes/no confirmation
func ConfirmAction(message string) bool {
	input, err := Prompt(message)
	if err != nil {
		return false
	}
	return input == "y" || input == "Y"
}

// ParseResourceName parses resource and returns full name with version and extension
func ParseResourceName(resource string) string {
	name, version, ext := store.ParseResourceName(resource)
	fullName := name
	if version != "" {
		fullName = name + "@" + version
	}
	if ext != "" {
		fullName += ext
	}
	return fullName
}

// FindMatchingResources returns all names in the list whose base name (and
// optional extension) match. Pass ext="" to match any extension.
func FindMatchingResources(all []string, baseName, ext string) []string {
	var matches []string
	for _, item := range all {
		itemName, _, itemExt := store.ParseResourceName(item)
		if itemName != baseName {
			continue
		}
		if ext != "" && itemExt != ext {
			continue
		}
		matches = append(matches, item)
	}
	return matches
}

// PickFromList shows a numbered list and returns the selected item.
// If only one item is in the list it is returned immediately without prompting.
func PickFromList(label string, items []string) (string, error) {
	if len(items) == 1 {
		return items[0], nil
	}
	fmt.Printf("Multiple versions found for '%s':\n", label)
	for i, name := range items {
		fmt.Printf("  %d. %s\n", i+1, name)
	}
	choice, err := Prompt(fmt.Sprintf("Enter version number (1-%d): ", len(items)))
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	choice = strings.TrimSpace(choice)
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(items) {
		return "", fmt.Errorf("invalid choice '%s'", choice)
	}
	return items[idx-1], nil
}
