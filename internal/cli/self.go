package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var selfCmd = &cobra.Command{
	Use:   "self",
	Short: "Manage Boiler installation",
	Long: `Manage Boiler CLI installation.

Commands for updating and uninstalling Boiler itself.`,
}

var selfUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Boiler CLI",
	Long: `Uninstall Boiler CLI from your system.

This will:
  - Remove the binary from installation directory
  - Clean PATH environment variable
  - Optionally remove config and store data

You will be prompted for confirmation before deletion.`,
	Example: `  # Uninstall Boiler
  bl self uninstall`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSelfUninstall(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var selfUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Boiler to latest version",
	Long: `Update Boiler CLI to the latest version.

Downloads and installs the latest release from GitHub.`,
	Example: `  # Update to latest version
  bl self update`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSelfUpdate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runSelfUninstall() error {
	
	var cmd *exec.Cmd
	scriptURL := "https://raw.githubusercontent.com/rishiyaduwanshi/boiler/main/scripts/uninstall"
	
	if runtime.GOOS == "windows" {
		scriptURL += ".ps1"
		// Download and execute uninstall script
		psCmd := fmt.Sprintf("irm %s | iex", scriptURL)
		cmd = exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psCmd)
	} else {
		scriptURL += ".sh"
		// Download and execute uninstall script
		bashCmd := fmt.Sprintf("curl -fsSL %s | bash", scriptURL)
		cmd = exec.Command("bash", "-c", bashCmd)
	}
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	return cmd.Run()
}

func runSelfUpdate() error {
	fmt.Println("Checking for updates...")
	
	var cmd *exec.Cmd
	installURL := "https://boiler.iamabhinav.dev/install"
	
	if runtime.GOOS == "windows" {
		// Use PowerShell to download and execute install script
		psCmd := fmt.Sprintf("iwr -useb %s | iex", installURL)
		cmd = exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psCmd)
	} else {
		// Use curl to download and execute install script
		bashCmd := fmt.Sprintf("curl -fsSL %s | bash", installURL)
		cmd = exec.Command("bash", "-c", bashCmd)
	}
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}
	
	fmt.Println("\n✓ Update complete! Restart your terminal.")
	return nil
}

func init() {
	selfCmd.AddCommand(selfUninstallCmd)
	selfCmd.AddCommand(selfUpdateCmd)
}
