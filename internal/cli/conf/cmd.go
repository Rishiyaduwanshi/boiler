package conf

import (
	"fmt"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"os"
	"os/exec"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "conf",
	Short: "Manage boiler configuration",
	Long: `View and manage Boiler configuration.

You can:
  - View current configuration (default)
  - Edit config in default editor (use -e or --edit)
  - Reset to defaults (use -r or --reset)
  - Set registry URL (use --set-registry)

Configuration includes paths, preferences, and behavior settings.`,
	Example: `  # Show configuration
  bl conf

  # Edit configuration (uses defaultEditor from config, or $EDITOR env)
  bl conf -e 

  # Edit with a specific editor
  bl conf -e code
  bl conf --edit notepad
  bl conf -e vim

  # Reset to defaults
  bl conf --reset

  # Set custom registry
  bl conf --set-registry https://github.com/myorg/boiler

  # Set to default registry
  bl conf --set-registry https://github.com/rishiyaduwanshi/boiler`,
	Run: func(cmd *cobra.Command, args []string) {
		if confSetRegistry != "" {
			if err := setRegistry(confSetRegistry); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
		if confEdit != "" {
			if err := editConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
		if confReset {
			fmt.Fprintf(os.Stderr, "Error: reset is not supported in the new architecture yet\n")
			os.Exit(1)
			return
		}
		showConfig()
	},
}

var (
	confEdit        string
	confReset       bool
	confSetRegistry string
)

func init() {
	Cmd.Flags().StringVarP(&confEdit, "edit", "e", "", "Edit configuration (optional: editor name)")
	// Cmd.Flags().Lookup("edit").NoOptDefVal = "__default__"
	Cmd.Flags().BoolVarP(&confReset, "reset", "r", false, "Reset configuration to defaults")
	Cmd.Flags().StringVar(&confSetRegistry, "set-registry", "", "Set custom registry URL")
}

func getActiveConfigPath() (string, error) {
	if config.Ctx.Scope == config.ScopeLocal {
		if config.Ctx.Manager.Local != nil && config.Ctx.Manager.Local.Path != "" {
			return config.Ctx.Manager.Local.Path, nil
		}
		return "", fmt.Errorf("local config path not found")
	}
	if config.Ctx.Manager.Global != nil && config.Ctx.Manager.Global.Path != "" {
		return config.Ctx.Manager.Global.Path, nil
	}
	return "", fmt.Errorf("global config path not found")
}

func showConfig() {
	configPath, err := getActiveConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting config path: %v\n", err)
		return
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		return
	}

	fmt.Printf("Config File: %s\n\n", configPath)
	fmt.Println(string(data))
}

func editConfig() error {
	configPath, err := getActiveConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Priority: explicit editor from -e flag → config defaultEditor
	var editor string
	if confEdit != "__default__" {
		editor = confEdit
	} else {
		editor = cfg.DefaultEditor
	}

	// Verify editor exists
	if _, err := exec.LookPath(editor); err != nil {
		return fmt.Errorf("editor %q not found. Set defaultEditor in boiler.conf.json or pass editor name: bl conf -e <editor>", editor)
	}

	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Info(fmt.Sprintf("Opening config in %s", editor))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	return nil
}

func setRegistry(registryURL string) error {
	// Validate URL format
	if registryURL == "" {
		return fmt.Errorf("registry URL cannot be empty")
	}

	if config.Ctx.Manager.Global == nil || config.Ctx.Manager.Global.Config == nil {
		return fmt.Errorf("global config not found")
	}

	// Update registry in global config (registry is inherently global)
	oldRegistry := config.Ctx.Manager.Global.Config.Registry
	config.Ctx.Manager.Global.Config.Registry = registryURL

	// Save global config
	if err := config.Ctx.Manager.SaveGlobal(); err != nil {
		return fmt.Errorf("failed to save global config: %w", err)
	}

	logger.Info(fmt.Sprintf("Registry changed from %s to %s", oldRegistry, registryURL))
	fmt.Printf("✓ Registry updated\n")
	fmt.Printf("  Old: %s\n", oldRegistry)
	fmt.Printf("  New: %s\n", registryURL)
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
