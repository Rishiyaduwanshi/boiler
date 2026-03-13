package cli

import (
	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var (
	cfg    *config.Config
	logger *utils.Logger
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "bl",
	Short: "Boiler - Code snippet and stack manager",
	Long: `Boiler - A CLI tool to manage reusable code snippets and project stacks.

Store, version, and reuse your code across projects. Perfect for:
  - Reusable utility functions (snippets)
  - Project templates and boilerplates (stacks)
  - Code patterns you use frequently
  - Multi-language development workflows

All resources are versioned automatically, making it easy to manage multiple
variations of the same snippet or stack.`,
	Example: `  # Initialize Boiler
  bl init

  # Store a snippet
  bl store ./utils/logger.js

  # Add snippet to project
  bl add logger

  # List all resources
  bl ls

  # Show paths
  bl path`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show banner when no subcommand is provided
		utils.ShowBanner()
		utils.ShowQuickHelp()
	},
}

// Execute runs the CLI
func Execute(config *config.Config, log *utils.Logger) error {
	cfg = config
	logger = log
	rootCmd.SilenceErrors = true
	return rootCmd.Execute()
}

// GetRootCommand returns the root command for documentation generation
func GetRootCommand() *cobra.Command {
	return rootCmd
}

func init() {
	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(confCmd)
	rootCmd.AddCommand(varCmd)
	rootCmd.AddCommand(unvarCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(storeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(pathCmd)
	rootCmd.AddCommand(selfCmd)
}
