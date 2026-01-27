package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show boiler installation path",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Boiler root: %s\n", cfg.Paths.Root)
		fmt.Printf("Store:       %s\n", cfg.Paths.Store)
		fmt.Printf("Snippets:    %s\n", cfg.Paths.Snippets)
		fmt.Printf("Stacks:      %s\n", cfg.Paths.Stacks)
		fmt.Printf("Logs:        %s\n", cfg.Paths.Logs)
		fmt.Printf("Bin:         %s\n", cfg.Paths.Bin)
	},
}
