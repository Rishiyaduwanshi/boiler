package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var varCmd = &cobra.Command{
	Use:   "var [name|name=value]",
	Short: "Manage reusable command and snippet variables",
	Long: `Manage reusable variables stored in boiler.conf.json.

Usage patterns:
  bl var                 List all variables
  bl var name=value      Set or update a variable
  bl var name            Get one variable value
  bl unvar name          Remove a variable (use 'unvar' command)

Variable names are normalized internally:
	- Optional prefixes : and bl__ are accepted
  - Names are case-insensitive
	- Hyphens and underscores are preserved

Examples:
  bl var API_URL=https://api.example.com
  bl var bl__TEAM_REG=https://github.com/myorg/boiler
  bl var api_url`,
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

func ensureConfigVars() {
	if cfg.Vars == nil {
		cfg.Vars = make(map[string]string)
		return
	}
	cfg.Vars = utils.NormalizeVarMap(cfg.Vars)
}

func persistConfigVars() error {
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
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

	ensureConfigVars()
	cfg.Vars[key] = rawValue

	if err := persistConfigVars(); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Variable set: %s", key))
	fmt.Printf("✓ Variable '%s' set\n", key)
	return nil
}

func getVarByName(rawKey string) error {
	ensureConfigVars()

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
	ensureConfigVars()
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
