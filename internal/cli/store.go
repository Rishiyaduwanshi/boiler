package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rishiyaduwanshi/boiler/internal/models"
	"github.com/rishiyaduwanshi/boiler/internal/store"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var (
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
)

var storeCmd = &cobra.Command{
	Use:   "store [path]",
	Short: "Store a folder/file as snippet or stack",
	Long: `Store a file as a snippet or directory as a stack in your Boiler store.

Files are stored as snippets with version numbers.
Directories must have a boiler.stack.json config file (run 'bl init' first).

Version Management:
  If snippet already exists, you'll be prompted with options:
    (o) Overwrite - Replace the latest version with new content
    (n) New version - Create a new incremental version
    (c) Cancel - Abort the operation
  First-time storage automatically creates version 1

Stacks require boiler.stack.json with:
  - id: Stack name
  - version: Version number
  - ignore: Patterns to exclude

If a stack version already exists, you'll be prompted to overwrite.`,
	Example: `  # Store current directory as stack
  bl store

  # Store specific file as snippet (first version)
  bl store ./utils/logger.js
  # Output: ✓ Stored snippet 'logger@1.js'

  # Store again - prompts for action
  bl store ./utils/logger.js
  # Prompt: Snippet 'logger' already exists (1 version(s)). Options:
  #   (o) Overwrite latest version (1)
  #   (n) Create new version (2)
  #   (c) Cancel

  # Store directory as stack
  bl store ./my-template

  # Store with custom name
  bl store ./config.js --name dbConfig.js`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		logger.Info(fmt.Sprintf("Storing: %s", path))

		if err := storeResource(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func storeResource(path string) error {
	// Check if path exists
	if !utils.FileExists(path) {
		return fmt.Errorf("path '%s' does not exist", path)
	}

	// Auto-detect type if not provided
	isDir := utils.IsDirectory(path)
	autoDetectedType := "snippet"
	if isDir {
		autoDetectedType = "stack"
	}

	// Use flags to override auto-detection
	if storeAsSnippet {
		autoDetectedType = "snippet"
	} else if storeAsStack {
		autoDetectedType = "stack"
	}

	// Get base name if not provided
	if storeName == "" {
		storeName = filepath.Base(path)
		// Remove extension for snippets to get clean name
		if !isDir {
			storeName = strings.TrimSuffix(storeName, filepath.Ext(storeName))
		}
	}

	st := store.NewStore(cfg.Paths.Store)
	if err := st.Load(); err != nil {
		return fmt.Errorf("failed to load store: %w", err)
	}

	if autoDetectedType == "snippet" {
		return storeSnippet(st, path)
	}
	return storeStack(st, path)
}

func storeSnippet(st *store.Store, path string) error {
	// Must be a file
	if utils.IsDirectory(path) {
		return fmt.Errorf("snippet must be a file, not a directory")
	}

	// Get extension
	ext := filepath.Ext(path)
	if ext == "" {
		return fmt.Errorf("snippet file must have an extension")
	}

	// Parse metadata from file comments
	meta, err := utils.ParseSnippetMetadata(path)
	if err != nil {
		return fmt.Errorf("failed to parse snippet metadata: %w", err)
	}

	// Validate required fields (only author needed)
	if err := utils.ValidateSnippetMetadata(meta); err != nil {
		return fmt.Errorf("invalid snippet metadata: %w\n\nAdd required metadata comment:\n  // __author Your Name", err)
	}

	// Use metadata name if no custom name provided
	if storeName == "" || storeName == strings.TrimSuffix(filepath.Base(path), ext) {
		if meta.Name != "" {
			storeName = meta.Name
		} else {
			storeName = strings.TrimSuffix(filepath.Base(path), ext)
		}
	}

	// Check if any version of this snippet exists
	existingVersions := st.GetAllVersions(storeName, ext)
	var version int
	var fullName string
	
	if len(existingVersions) > 0 {
		// Snippet already exists
		latestVersion := existingVersions[len(existingVersions)-1]
		
		choice, err := utils.Prompt(fmt.Sprintf(
			"Snippet '%s' already exists (%d version(s)). Options:\n  (o) Overwrite latest version (%d)\n  (n) Create new version (%d)\n  (c) Cancel\nChoice: ",
			storeName+ext, len(existingVersions), latestVersion, latestVersion+1))
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		
		choice = strings.ToLower(strings.TrimSpace(choice))
		switch choice {
		case "o", "overwrite":
			// Overwrite latest version
			version = latestVersion
		case "n", "new":
			// Create new version
			version = latestVersion + 1
		case "c", "cancel":
			return fmt.Errorf("cancelled by user")
		default:
			return fmt.Errorf("invalid choice '%s'. Use 'o' for overwrite, 'n' for new version, or 'c' to cancel", choice)
		}
	} else {
		// First version
		version = 1
	}

	// Determine the language directory based on extension
	langDir := strings.TrimPrefix(ext, ".")
	snippetDir := filepath.Join(cfg.Paths.Snippets, langDir)
	if err := utils.EnsureDir(snippetDir); err != nil {
		return fmt.Errorf("failed to create snippet directory: %w", err)
	}

	// Build full name with version
	fullName = fmt.Sprintf("%s@%d%s", storeName, version, ext)
	destPath := filepath.Join(snippetDir, filepath.Base(fullName))

	// If overwriting, remove old version
	if st.SnippetExists(fullName) {
		if err := st.RemoveSnippet(fullName); err != nil {
			return fmt.Errorf("failed to remove old snippet: %w", err)
		}
		if utils.FileExists(destPath) {
			if err := os.Remove(destPath); err != nil {
				return fmt.Errorf("failed to remove old file: %w", err)
			}
		}
	}

	// Copy file
	if err := utils.CopyFile(path, destPath); err != nil {
		return fmt.Errorf("failed to copy snippet: %w", err)
	}

	// Add to metadata
	if err := st.AddSnippet(fullName, destPath); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	fmt.Printf("✓ Stored snippet '%s' at %s\n", fullName, destPath)
	logger.Info(fmt.Sprintf("Snippet stored: %s -> %s", path, destPath))
	return nil
}

func storeStack(st *store.Store, path string) error {
	// Must be a directory
	if !utils.IsDirectory(path) {
		return fmt.Errorf("stack must be a directory, not a file")
	}

	// Parse config (mandatory)
	stackConfig, err := models.ParseStackConfig(path)
	if err != nil {
		return err
	}

	// Validate required fields
	if stackConfig.ID == "" {
		return fmt.Errorf("'id' field is required in boiler.stack.json")
	}
	if stackConfig.Version == "" {
		return fmt.Errorf("'version' field is required in boiler.stack.json")
	}

	// Use config ID as stack name
	storeName = stackConfig.ID

	// Parse version
	version, err := strconv.Atoi(stackConfig.Version)
	if err != nil {
		return fmt.Errorf("invalid version in boiler.stack.json: %s", stackConfig.Version)
	}

	// Build paths
	fullName := fmt.Sprintf("%s@%d", storeName, version)
	stackDir := filepath.Join(cfg.Paths.Stacks, fullName)

	// Check if this version already exists
	if st.StackExists(fullName) {
		choice, err := utils.Prompt(fmt.Sprintf("Stack '%s' already exists. Overwrite? (y/n): ", fullName))
		if err != nil || strings.ToLower(strings.TrimSpace(choice)) != "y" {
			return fmt.Errorf("cancelled")
		}
		// Remove old version
		if err := st.RemoveStack(fullName); err != nil {
			return fmt.Errorf("failed to remove old stack: %w", err)
		}
		if utils.IsDirectory(stackDir) {
			if err := os.RemoveAll(stackDir); err != nil {
				return fmt.Errorf("failed to remove old directory: %w", err)
			}
		}
	}

	// Get ignore patterns from config
	ignorePatterns := models.ResolveIgnorePatterns(stackConfig)

	// Copy directory
	if err := utils.CopyDir(path, stackDir, ignorePatterns); err != nil {
		return fmt.Errorf("failed to copy stack: %w", err)
	}

	// Add to metadata
	if err := st.AddStack(fullName, stackDir); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	fmt.Printf("✓ Stored stack '%s' at %s\n", fullName, stackDir)
	logger.Info(fmt.Sprintf("Stack stored: %s -> %s", path, stackDir))
	return nil
}

var (
	storeName      string
	storeAsSnippet bool
	storeAsStack   bool
)

func init() {
	storeCmd.Flags().StringVarP(&storeName, "name", "", "", "Name for the resource (auto-detected from path if not provided)")
	storeCmd.Flags().BoolVarP(&storeAsSnippet, "snippet", "n", false, "Force store as snippet")
	storeCmd.Flags().BoolVarP(&storeAsStack, "stack", "k", false, "Force store as stack")
}
