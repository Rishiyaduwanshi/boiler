package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/spf13/cobra"
)

var confCmd = &cobra.Command{
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

  # Edit configuration
  bl conf --edit

  # Reset to defaults
  bl conf --reset

  # Set custom registry
  bl conf --set-registry https://github.com/myorg/boiler

  # Set to default registry
  bl conf --set-registry https://github.com/rishiyaduwanshi/boiler`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior: show config
		showConfig()
	},
}

var (
	confEdit        bool
	confReset       bool
	confShow        bool
	confSetRegistry string
)

func init() {
	confCmd.Flags().BoolVarP(&confEdit, "edit", "e", false, "Edit configuration")
	confCmd.Flags().BoolVarP(&confReset, "reset", "r", false, "Reset configuration to defaults")
	confCmd.Flags().BoolVarP(&confShow, "show", "s", false, "Show configuration")
	confCmd.Flags().StringVar(&confSetRegistry, "set-registry", "", "Set custom registry URL")

	// Set PreRunE to handle edit and reset flags
	confCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if confSetRegistry != "" {
			if err := setRegistry(confSetRegistry); err != nil {
				return err
			}
			os.Exit(0)
		}
		if confEdit {
			if err := editConfig(); err != nil {
				return err
			}
			os.Exit(0) // Exit after editing
		}
		if confReset {
			if err := resetConfig(); err != nil {
				return err
			}
			os.Exit(0) // Exit after reset
		}
		return nil
	}
}

func showConfig() {
	configPath, err := config.ConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting config path: %v\n", err)
		return
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		return
	}

	fmt.Println(string(data))
}

func editConfig() error {
	configPath, err := config.ConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	editor := cfg.DefaultEditor
	if envEditor := os.Getenv("EDITOR"); envEditor != "" {
		editor = envEditor
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

func resetConfig() error {
	logger.Info("Resetting configuration to defaults")

	if err := config.Reset(); err != nil {
		return fmt.Errorf("failed to reset config: %w", err)
	}

	fmt.Println("Configuration reset to defaults")
	return nil
}

func setRegistry(registryURL string) error {
	// Validate URL format
	if registryURL == "" {
		return fmt.Errorf("registry URL cannot be empty")
	}

	// Load current config
	currentCfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Update registry
	oldRegistry := currentCfg.Registry
	currentCfg.Registry = registryURL

	// Save config
	if err := config.Save(currentCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	logger.Info(fmt.Sprintf("Registry changed from %s to %s", oldRegistry, registryURL))
	fmt.Printf("✓ Registry updated\n")
	fmt.Printf("  Old: %s\n", oldRegistry)
	fmt.Printf("  New: %s\n", registryURL)
	return nil
}
