package unalias

import (
	"fmt"
	aliascmd "github.com/rishiyaduwanshi/boiler/internal/cli/alias"
	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"os"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
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
	aliascmd.EnsureConfigAliases()

	name, err := aliascmd.NormalizeAliasName(rawName)
	if err != nil {
		return err
	}

	if _, ok := cfg.Aliases[name]; !ok {
		return fmt.Errorf("alias '%s' not found", name)
	}

	delete(cfg.Aliases, name)

	// Remove alias from the scope-owned config object before saving to disk
	if err := aliascmd.DeleteScopedAlias(name); err != nil {
		return err
	}

	if err := aliascmd.PersistConfigAliases(); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Alias removed: %s", name))
	fmt.Printf("✓ Alias '%s' removed\n", name)
	return nil
}

var (
	cfg    *config.Config
	logger *utils.Logger
)

func Setup(c *config.Config, l *utils.Logger) {
	cfg = c
	logger = l
}
