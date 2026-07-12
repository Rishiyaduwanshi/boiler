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

		// Find project root automatically like npm/mise
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		projectRoot := cwd
		nearestConfig, err := config.FindNearestConfig(cwd)
		if err == nil && nearestConfig != "" {
			projectRoot = filepath.Dir(nearestConfig)
		}

		// In Boiler, local scripts are in the `bl/commands/` directory of the project
		scriptPath := filepath.Join(projectRoot, "bl", "commands", scriptName+".bl")

		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return fmt.Errorf("script '%s' not found. Looked for it at: %s", scriptName, scriptPath)
		}

		// Parse dynamic arguments and flags
		vars := make(map[string]string)
		positionalIndex := 1

		// Seed vars from BOILER_VAR_* environment variables so that env-prefilled
		// values are available during direct .bl script execution (create, inject, etc.)
		// CLI args added below will override these if the same key is provided.
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, "BOILER_VAR_") {
				kv := strings.SplitN(env, "=", 2)
				if len(kv) == 2 {
					key := strings.ToLower(strings.TrimPrefix(kv[0], "BOILER_VAR_"))
					vars[key] = kv[1]
				}
			}
		}

		for i := 1; i < len(args); i++ {
			arg := args[i]

			// Global flags shouldn't be passed to the script variables ideally, but for now we consume all.
			if arg == "--verbose" || arg == "-V" {
				continue // skip global verbose flag
			}

			if arg == "--global" || arg == "--local" {
				continue
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
			logger.Info(fmt.Sprintf("Executing Boiler script: %s from %s", scriptPath, projectRoot))
		} else {
			fmt.Printf("Executing Boiler script: %s from %s\n", scriptPath, projectRoot)
		}

		// Change directory to project root before execution so relative paths work properly.
		// Both the switch and the restore must succeed to prevent files landing in the wrong tree.
		originalDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to determine current directory: %w", err)
		}
		if err := os.Chdir(projectRoot); err != nil {
			return fmt.Errorf("failed to switch to project root %q: %w", projectRoot, err)
		}
		defer func() {
			if restoreErr := os.Chdir(originalDir); restoreErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to restore working directory to %q: %v\n", originalDir, restoreErr)
			}
		}()

		err = engine.ParseAndExecute(scriptPath, vars)
		if err != nil {
			return fmt.Errorf("script execution failed: %w", err)
		}

		if logger != nil {
			logger.Info(fmt.Sprintf("Successfully executed '%s'", scriptName))
		}
		return nil
	},
}
