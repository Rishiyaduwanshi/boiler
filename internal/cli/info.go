package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [resource]",
	Short: "Show detailed information about a resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resource := args[0]
		logger.Info(fmt.Sprintf("Getting info for: %s", resource))
		fmt.Printf("Info for %s (implementation in progress)\n", resource)
	},
}
