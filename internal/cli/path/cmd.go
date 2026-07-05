package path

import (
	"fmt"
	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "path",
	Short: "Show boiler installation path",
	Long: `Display all Boiler installation paths.

Shows:
  - Root - Main Boiler directory
  - Store - Where resources are stored
  - Snippets - Snippet storage location
  - Stacks - Stack storage location
  - Logs - Log file directory
  - Bin - Executable location`,
	Example: `  # Show all paths
  bl path

  # Use in scripts
  cd $(bl path | grep Store | cut -d: -f2)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Boiler root: %s\n", cfg.Paths.Root)
		fmt.Printf("Store:       %s\n", cfg.Paths.Store)
		fmt.Printf("Snippets:    %s\n", cfg.Paths.Snippets)
		fmt.Printf("Stacks:      %s\n", cfg.Paths.Stacks)
		fmt.Printf("Logs:        %s\n", cfg.Paths.Logs)
		fmt.Printf("Bin:         %s\n", cfg.Paths.Bin)
	},
}

var (
	cfg    *config.Config
	logger *utils.Logger
)

func Setup(c *config.Config, l *utils.Logger) {
	cfg = c
	logger = l
}
