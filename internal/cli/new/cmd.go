package new

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/engine"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var (
	cfg    *config.Config
	logger *utils.Logger
)

// Setup injects the configuration and logger
func Setup(c *config.Config, l *utils.Logger) {
	cfg = c
	logger = l
}

// Cmd represents the 'new' command
var Cmd = &cobra.Command{
	Use:   "new [script_name] [args...]",
	Short: "Generate code using a Boiler script (.bl)",
	Long: `Run a Boiler script (.bl) to generate, inject, or modify code in your project.
	
Boiler automatically parses any flags (like --ts or --port=3000) and maps them to script variables.
It looks for the script in the './bl/' folder of your current project.`,
	Example: `  # Run the routes.bl script with positional arguments
  bl new routes user auth

  # Run with flags
  bl new routes --ts --port=3000`,
	DisableFlagParsing: true, // Crucial: allows dynamic user-defined flags like --ts
	RunE: func(cmd *cobra.Command, args []string) error {
		// Because DisableFlagParsing is true, the very first arg might be the global flag
		// if passed after 'new' (e.g. bl new --help). We should handle it if needed.
		if len(args) == 0 {
			return fmt.Errorf("please provide a script name (e.g., 'bl new routes')")
		}

		if args[0] == "-h" || args[0] == "--help" {
			return cmd.Help()
		}

		scriptName := args[0]
		// In Boiler, local scripts are in the `bl/` directory of the project
		scriptPath := filepath.Join("bl", scriptName+".bl")

		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return fmt.Errorf("script '%s' not found. Ensure '%s' exists in your project", scriptName, scriptPath)
		}

		// Parse dynamic arguments and flags
		vars := make(map[string]string)
		positionalIndex := 1

		for i := 1; i < len(args); i++ {
			arg := args[i]

			// Global flags shouldn't be passed to the script variables ideally, but for now we consume all.
			if arg == "--verbose" || arg == "-V" {
				continue // skip global verbose flag
			}

			if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
				// It's a flag
				flagRaw := strings.TrimLeft(arg, "-")
				if strings.Contains(flagRaw, "=") {
					parts := strings.SplitN(flagRaw, "=", 2)
					vars["bl__"+parts[0]] = parts[1]
				} else {
					vars["bl__"+flagRaw] = "true"
				}
			} else {
				// It's a positional argument
				vars[fmt.Sprintf("bl__%d", positionalIndex)] = arg
				positionalIndex++
			}
		}

		if logger != nil {
			logger.Info(fmt.Sprintf("Executing Boiler script: %s", scriptPath))
		} else {
			fmt.Printf("Executing Boiler script: %s\n", scriptPath)
		}

		err := engine.ParseAndExecute(scriptPath, vars)
		if err != nil {
			return fmt.Errorf("script execution failed: %w", err)
		}

		if logger != nil {
			logger.Info(fmt.Sprintf("Successfully executed '%s'", scriptName))
		}
		return nil
	},
}
