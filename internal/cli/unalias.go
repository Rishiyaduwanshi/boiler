package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var unaliasCmd = &cobra.Command{
	Use:   "unalias [name]",
	Short: "Remove a command alias",
	Long: `Remove a command alias from boiler.conf.json.

Examples:
  bl unalias ll
  bl unalias search_shortcut`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := unsetAlias(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func unsetAlias(rawName string) error {
	ensureConfigAliases()

	name, err := normalizeAliasName(rawName)
	if err != nil {
		return err
	}

	if _, ok := cfg.Aliases[name]; !ok {
		return fmt.Errorf("alias '%s' not found", name)
	}

	delete(cfg.Aliases, name)
	if err := persistConfigAliases(); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Alias removed: %s", name))
	fmt.Printf("✓ Alias '%s' removed\n", name)
	return nil
}
