package cli

import (
	"fmt"
	"os"

	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var unvarCmd = &cobra.Command{
	Use:   "unvar [name]",
	Short: "Remove a variable",
	Long: `Remove a variable from boiler.conf.json.

Examples:
  bl unvar API_URL
  bl unvar bl__TEAM_REG
  bl unvar @team_reg`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := unsetVar(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func unsetVar(rawKey string) error {
	ensureConfigVars()

	key, err := utils.NormalizeVarKey(rawKey)
	if err != nil {
		return err
	}

	if _, ok := cfg.Vars[key]; !ok {
		return fmt.Errorf("variable '%s' not found", key)
	}

	delete(cfg.Vars, key)
	if err := persistConfigVars(); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Variable removed: %s", key))
	fmt.Printf("✓ Variable '%s' removed\n", key)
	return nil
}
