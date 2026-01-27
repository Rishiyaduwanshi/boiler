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
	Long:  `A CLI tool to manage reusable code snippets and stacks across multiple programming languages.`,
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
	return rootCmd.Execute()
}

func init() {
	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(confCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(storeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(pathCmd)
}
