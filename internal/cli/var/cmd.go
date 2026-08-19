package varcmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "var [name|name=value]",
	Short: "Manage reusable command and snippet variables",
	Long: `Manage reusable variables stored in boiler.conf.json.

Usage patterns:
  bl var                       List all variables
  bl var name=value            Set a new variable
  bl var name=value --force    Overwrite an existing variable
  bl var name                  Get one variable value
  bl unvar name                Remove a variable (use 'unvar' command)

Variables can be used inline in ANY command using the bl__VAR_NAME syntax:
  - When Boiler encounters bl__VAR_NAME in an argument, it replaces it with the variable's value.
  - Useful for dynamic paths, URLs, and registry configurations.

Examples:
  # Set variables
  bl var org=rishiyaduwanshi
  bl var repo=boiler-templates

  # Use variables inline in a command (fetches github.com/rishiyaduwanshi/boiler-templates:auth)
  bl use github.com/bl__org/bl__repo:auth`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		switch len(args) {
		case 0:
			err = listVars()
		default:
			input := strings.TrimSpace(args[0])
			if strings.Contains(input, "=") {
				err = setVarFromAssignment(input)
			} else {
				err = getVarByName(input)
			}
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}
var forceOverwrite bool

func init() {
	Cmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "Overwrite if already exists")
}

func EnsureConfigVars() {
	if cfg.Vars == nil {
		cfg.Vars = make(map[string]string)
		return
	}
	cfg.Vars = utils.NormalizeVarMap(cfg.Vars)
}

func PersistConfigVars() error {
	if config.Ctx.Scope == config.ScopeLocal {
		if err := config.Ctx.Manager.SaveLocal(); err != nil {
			return fmt.Errorf("failed to save local config: %w", err)
		}
	} else {
		if err := config.Ctx.Manager.SaveGlobal(); err != nil {
			return fmt.Errorf("failed to save global config: %w", err)
		}
	}
	return nil
}

func setVarFromAssignment(assignment string) error {
	rawKey, rawValue, found := strings.Cut(assignment, "=")
	if !found {
		return fmt.Errorf("invalid variable assignment; use KEY=VALUE")
	}

	key, err := utils.NormalizeVarKey(rawKey)
	if err != nil {
		return err
	}

	EnsureConfigVars()

	if _, ok := config.ScopedVarMap()[key]; ok && !forceOverwrite {
		return fmt.Errorf("'%s' already exists. Use --force to overwrite", key)
	}

	cfg.Vars[key] = rawValue

	if err := config.SetScopedVar(key, rawValue); err != nil {
		return err
	}

	if err := config.PersistScopedVars(); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Variable set: %s", key))
	fmt.Printf("✓ Variable '%s' set\n", key)
	return nil
}

func getVarByName(rawKey string) error {
	EnsureConfigVars()

	key, err := utils.NormalizeVarKey(rawKey)
	if err != nil {
		return err
	}

	value, ok := cfg.Vars[key]
	if !ok {
		return fmt.Errorf("variable '%s' not found", key)
	}

	fmt.Printf("%s=%s\n", key, value)
	return nil
}

func listVars() error {
	EnsureConfigVars()
	if len(cfg.Vars) == 0 {
		fmt.Println("No variables configured")
		return nil
	}

	keys := make([]string, 0, len(cfg.Vars))
	for key := range cfg.Vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fmt.Println("Variables:")
	for _, key := range keys {
		fmt.Printf("  %s=%s\n", key, cfg.Vars[key])
	}
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
