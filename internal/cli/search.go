package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for snippets or stacks",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		logger.Info(fmt.Sprintf("Searching for: %s", query))
		fmt.Printf("Searching for %s (implementation in progress)\n", query)
	},
}

var (
	searchSnippets bool
	searchStacks   bool
	searchRemote   bool
)

func init() {
	searchCmd.Flags().BoolVarP(&searchSnippets, "snippets", "n", false, "Search only snippets")
	searchCmd.Flags().BoolVarP(&searchStacks, "stacks", "k", false, "Search only stacks")
	searchCmd.Flags().BoolVarP(&searchRemote, "remote", "r", false, "Search remote registry")
}
