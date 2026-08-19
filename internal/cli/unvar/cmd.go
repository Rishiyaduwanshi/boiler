package unvar

import (
	"fmt"
	varcmd "github.com/rishiyaduwanshi/boiler/internal/cli/var"
	"github.com/rishiyaduwanshi/boiler/internal/config"
	"os"

	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "unvar [name]",
	Short: "Remove a variable",
	Long: `Remove a variable from boiler.conf.json.

Examples:
  bl unvar API_URL
  bl unvar bl__TEAM_REG
	bl unvar :team_reg`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := unsetVar(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func unsetVar(rawKey string) error {
	varcmd.EnsureConfigVars()

	key, err := utils.NormalizeVarKey(rawKey)
	if err != nil {
		return err
	}

	if _, ok := config.ScopedVarMap()[key]; !ok {
		return fmt.Errorf("'%s' not found in %s scope", key, config.Ctx.Scope)
	}

	delete(cfg.Vars, key)

	if err := config.DeleteScopedVar(key); err != nil {
		return err
	}

	if err := config.PersistScopedVars(); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Variable removed: %s", key))
	fmt.Printf("✓ Variable '%s' removed\n", key)
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
