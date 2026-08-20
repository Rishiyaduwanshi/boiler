package store

import (
	"fmt"
	"github.com/rishiyaduwanshi/boiler/internal/config"
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

var Cmd = &cobra.Command{
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

If a stack version already exists, you'll be prompted to overwrite.

Use --command (or --cmd) to store a .bl script in the global commands directory
(~/.boiler/commands/) so it can be run with 'bl new <script_name>'.`,
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
  bl store ./config.js --name dbConfig.js

  # Store a .bl script as a runnable command
  bl store ./route.bl --command`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		logger.Info(fmt.Sprintf("Storing: %s", path))

		if storeAsCommand {
			if err := storeCommand(path); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

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

	// Auto-detect type if not specified via flags
	isDir := utils.IsDirectory(path)
	autoDetectedType := "snippet"
	if isDir {
		autoDetectedType = "stack"
	}

	if storeAsSnippet {
		autoDetectedType = "snippet"
	} else if storeAsStack {
		autoDetectedType = "stack"
	}

	// Derive a local name from the flag; don't mutate the package-level var
	name := storeName
	if name == "" {
		name = filepath.Base(path)
		if !isDir {
			name = strings.TrimSuffix(name, filepath.Ext(name))
		}
	}

	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}

	if autoDetectedType == "snippet" {
		return storeSnippet(st, path, name)
	}
	return storeStack(st, path, name)
}

func storeSnippet(st *store.Store, path, name string) error {
	// Must be a file
	if utils.IsDirectory(path) {
		return fmt.Errorf("snippet must be a file, not a directory")
	}

	ext := filepath.Ext(path)
	if ext == "" {
		return fmt.Errorf("snippet file must have an extension")
	}

	meta, err := utils.ParseSnippetMetadata(path)
	if err != nil {
		return fmt.Errorf("failed to parse snippet metadata: %w", err)
	}

	if err := utils.ValidateSnippetMetadata(meta); err != nil {
		return fmt.Errorf("invalid snippet metadata: %w", err)
	}

	// Prefer metadata name, then fall back to the passed-in name
	if name == "" || name == strings.TrimSuffix(filepath.Base(path), ext) {
		if meta.Name != "" {
			name = meta.Name
		}
	}

	// Check if any version of this snippet exists
	existingVersions := st.GetAllVersions(name, ext)
	var version int
	var fullName string

	if len(existingVersions) > 0 {
		latestVersion := existingVersions[len(existingVersions)-1]

		choice, err := utils.Prompt(fmt.Sprintf(
			"Snippet '%s' already exists (%d version(s)). Options:\n  (o) Overwrite latest version (%d)\n  (n) Create new version (%d)\n  (c) Cancel\nChoice: ",
			name+ext, len(existingVersions), latestVersion, latestVersion+1))
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		choice = strings.ToLower(strings.TrimSpace(choice))
		switch choice {
		case "o", "overwrite":
			version = latestVersion
		case "n", "new":
			version = latestVersion + 1
		case "c", "cancel":
			return fmt.Errorf("cancelled by user")
		default:
			return fmt.Errorf("invalid choice '%s'. Use 'o' for overwrite, 'n' for new version, or 'c' to cancel", choice)
		}
	} else {
		version = 1
	}

	langDir := strings.TrimPrefix(ext, ".")
	snippetDir := filepath.Join(cfg.Paths.Snippets, langDir)
	if err := utils.EnsureDir(snippetDir); err != nil {
		return fmt.Errorf("failed to create snippet directory: %w", err)
	}

	fullName = fmt.Sprintf("%s@%d%s", name, version, ext)
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

func storeStack(st *store.Store, path, name string) error {
	// Must be a directory
	if !utils.IsDirectory(path) {
		return fmt.Errorf("stack must be a directory, not a file")
	}

	if storeName != "" {
		return fmt.Errorf("--name is not supported for stacks; set the stack name via the \"id\" field in boiler.stack.json")
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

	stackName := stackConfig.ID

	// Parse version
	version, err := strconv.Atoi(stackConfig.Version)
	if err != nil {
		return fmt.Errorf("invalid version in boiler.stack.json: %s", stackConfig.Version)
	}

	fullName := fmt.Sprintf("%s@%d", stackName, version)
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
	storeAsCommand bool
)

func init() {
	Cmd.Flags().StringVarP(&storeName, "name", "m", "", "Name for the resource (auto-detected from path if not provided)")
	Cmd.Flags().BoolVarP(&storeAsSnippet, "snippet", "n", false, "Force store as snippet")
	Cmd.Flags().BoolVarP(&storeAsStack, "stack", "k", false, "Force store as stack")
	Cmd.Flags().BoolVar(&storeAsCommand, "command", false, "Store .bl file in the global commands directory (~/.boiler/commands/)")
	Cmd.Flags().BoolVar(&storeAsCommand, "cmd", false, "Alias for --command")
}

var (
	cfg    *config.Config
	logger *utils.Logger
)

func Setup(c *config.Config, l *utils.Logger) {
	cfg = c
	logger = l
}

// storeCommand copies a .bl script to ~/.boiler/commands/ so it can be run via 'bl new <name>'.
// It always targets the global commands directory regardless of the current scope.
func storeCommand(path string) error {
	if !utils.FileExists(path) {
		return fmt.Errorf("file '%s' does not exist", path)
	}
	if utils.IsDirectory(path) {
		return fmt.Errorf("--command expects a file, not a directory")
	}
	if filepath.Ext(path) != ".bl" {
		return fmt.Errorf("--command only accepts .bl files, got '%s'", filepath.Base(path))
	}

	if err := utils.EnsureDir(cfg.Paths.Commands); err != nil {
		return fmt.Errorf("failed to create commands directory: %w", err)
	}

	destPath := filepath.Join(cfg.Paths.Commands, filepath.Base(path))
	if err := utils.CopyFile(path, destPath); err != nil {
		return fmt.Errorf("failed to copy command: %w", err)
	}

	fmt.Printf("✓ Stored command '%s' at %s\n", filepath.Base(path), destPath)
	logger.Info(fmt.Sprintf("Command stored: %s -> %s", path, destPath))
	return nil
}
