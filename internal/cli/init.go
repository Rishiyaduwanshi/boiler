package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/models"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize stack config in current directory",
	Long: `Initialize a boiler configuration file in the current directory.

For stacks (directories): Creates boiler.stack.json
For snippets (files): Creates boiler.snippet.json with metadata

Stack config includes:
  - Stack name and description
  - Author information
  - Files/folders to ignore
  - Version metadata

Snippet config includes:
  - Name, description, author
  - Language and tags
  - Version for templating

Similar to 'npm init', this helps you prepare projects for storing.`,
	Example: `  # Interactive init (prompts for details)
  bl init

  # Quick init with defaults (stack)
  bl init -y

  # Initialize snippet
  bl init --snippet
  bl init -n -y

  # After init, customize and store
  bl store`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := initBoilerConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}


func initBoilerConfig() error {
	// Determine what to initialize
	var isSnippet bool
	
	if initAsSnippet {
		// Explicitly requested snippet with -n flag
		isSnippet = true
	} else if initYes {
		// -y defaults to stack
		isSnippet = false
	} else {
		// Interactive: Ask user
		choice, err := utils.Prompt("Initialize as stack (k) or snippet (n)? ")
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		choice = strings.ToLower(strings.TrimSpace(choice))
		
		// Validate choice
		if choice != "k" && choice != "n" && choice != "stack" && choice != "snippet" {
			return fmt.Errorf("invalid choice '%s'. Please enter 'k' for stack or 'n' for snippet", choice)
		}
		
		isSnippet = (choice == "n" || choice == "snippet")
	}
	
	if isSnippet {
		// For snippet, just create file (will prompt for name)
		return createSnippetConfig("")
	}
	
	// For stack, check if config exists
	stackConfigPath := "./boiler.stack.json"
	if utils.FileExists(stackConfigPath) {
		return fmt.Errorf("boiler.stack.json already exists. Edit it or delete to reinitialize")
	}
	return createStackConfig(stackConfigPath)
}

func createStackConfig(path string) error {
	// Get current directory name as default
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defaultName := filepath.Base(cwd)
	
	// Prompt for common metadata
	commonMeta, err := utils.PromptCommonMetadata(defaultName, initYes)
	if err != nil {
		return err
	}
	
	config := models.StackConfig{
		ID:          commonMeta.Name,
		Version:     commonMeta.Version,
		Author:      commonMeta.Author,
		Description: commonMeta.Description,
		CreatedAt:   commonMeta.CreatedAt,
		Ignore:      []string{},
	}

	// Write to file
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Println("✓ Created boiler.stack.json")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit boiler.stack.json to customize settings")
	fmt.Println("  2. Run 'bl store' to save this stack")
	
	return nil
}

func createSnippetConfig(path string) error {
	// Ask for filename first
	var fileName string
	if initYes {
		fileName = "snippet.bl"
	} else {
		fileName = utils.PromptString("Filename (e.g., handler.js, Dockerfile, .gitignore)", "snippet.bl")
	}
	
	fileName = strings.TrimSpace(fileName)
	
	// Extract extension from filename
	ext := filepath.Ext(fileName)
	var artifact string
	var baseName string
	
	if ext != "" {
		// Has extension: handler.js → artifact="js", base="handler"
		artifact = strings.TrimPrefix(ext, ".")
		baseName = strings.TrimSuffix(fileName, ext)
	} else {
		// No extension: Dockerfile, Makefile
		// Ask for artifact to determine comment style
		if initYes {
			artifact = "default"
		} else {
			artifact = utils.PromptString("Artifact type (for comment style, e.g., dockerfile, gitignore)", "default")
		}
		baseName = fileName
	}
	
	// Get comment style from config, fallback to default if not found
	commentPrefix := cfg.Artifacts[artifact]
	if commentPrefix == "" {
		commentPrefix = cfg.Artifacts["default"]
	}
	
	// Prompt for metadata using base name as default
	commonMeta, err := utils.PromptCommonMetadata(baseName, initYes)
	if err != nil {
		return err
	}
	
	// Create snippet file with metadata comments
	if err := utils.GenerateSnippetTemplate(fileName, commonMeta, commentPrefix); err != nil {
		return err
	}
	
	fmt.Printf("✓ Created %s\n", fileName)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit the file and add your code")
	fmt.Println("  2. Add template variables: // __var bl__VAR_NAME = DefaultValue")
	fmt.Println("  3. Run 'bl store " + fileName + "' to save the snippet")
	
	return nil
}



var (
	initYes       bool
	initAsSnippet bool
)

func init() {
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "Skip prompts and use defaults")
	initCmd.Flags().BoolVarP(&initAsSnippet, "snippet", "n", false, "Initialize as snippet")
}
