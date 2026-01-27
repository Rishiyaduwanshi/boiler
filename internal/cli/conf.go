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
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior: show config
		showConfig()
	},
}

var (
	confEdit  bool
	confReset bool
	confShow  bool
)

func init() {
	confCmd.Flags().BoolVarP(&confEdit, "edit", "e", false, "Edit configuration")
	confCmd.Flags().BoolVarP(&confReset, "reset", "r", false, "Reset configuration to defaults")
	confCmd.Flags().BoolVarP(&confShow, "show", "s", false, "Show configuration")

	// Set PreRunE to handle edit and reset flags
	confCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
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
