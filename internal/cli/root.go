package cli

import (
	"os"
	"sync"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/remote"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var (
	cfg    *config.Config
	logger *utils.Logger
	verbose bool
)

var normalizeCommandDocsOnce sync.Once

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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if logger != nil {
			logger.SetVerbose(verbose)
		}
		remote.SetVerbose(verbose)
	},
}

// Execute runs the CLI
func Execute(config *config.Config, log *utils.Logger) error {
	cfg = config
	logger = log
	rootCmd.SilenceErrors = true
	rootCmd.SetArgs(expandFirstCommandAlias(os.Args[1:]))
	ensureCommandDocsNormalized()
	return rootCmd.Execute()
}

// GetRootCommand returns the root command for documentation generation
func GetRootCommand() *cobra.Command {
	ensureCommandDocsNormalized()
	return rootCmd
}

func ensureCommandDocsNormalized() {
	normalizeCommandDocsOnce.Do(func() {
		utils.NormalizeCommandTreeDocs(rootCmd)
	})
}

func init() {
	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "Enable verbose debug output")
	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(confCmd)
	rootCmd.AddCommand(aliasCmd)
	rootCmd.AddCommand(unaliasCmd)
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
