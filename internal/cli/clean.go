package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean [resource]",
	Short: "Clean snippets, stacks, or store",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			resource := args[0]
			logger.Info(fmt.Sprintf("Cleaning resource: %s", resource))
			fmt.Printf("Cleaning %s (implementation in progress)\n", resource)
			return
		}

		logger.Info("Starting interactive clean")
		fmt.Println("Interactive clean (implementation in progress)")
	},
}

var (
	cleanAll      bool
	cleanSnippets bool
	cleanStacks   bool
	cleanStore    bool
)

func init() {
	cleanCmd.Flags().BoolVarP(&cleanAll, "all", "a", false, "Clean everything")
	cleanCmd.Flags().BoolVarP(&cleanSnippets, "snippets", "n", false, "Clean all snippets")
	cleanCmd.Flags().BoolVarP(&cleanStacks, "stacks", "k", false, "Clean all stacks")
	cleanCmd.Flags().BoolVarP(&cleanStore, "store", "s", false, "Clean entire store")
}
