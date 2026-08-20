package new

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/constants"
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

// resolveScriptPath returns the path of the .bl script to execute based on the current scope.
func resolveScriptPath(scriptName, localCommandsDir, globalCommandsDir string, scope config.Scope, forceLocalOnly bool) (string, error) {
	local := filepath.Join(localCommandsDir, scriptName+".bl")
	global := filepath.Join(globalCommandsDir, scriptName+".bl")

	switch scope {
	case config.ScopeLocal:
		if _, err := os.Stat(local); err == nil {
			return local, nil
		}
		if forceLocalOnly {
			return "", fmt.Errorf("script '%s' not found in local commands directory: %s", scriptName, local)
		}
		// Fallback to global
		if _, err := os.Stat(global); err == nil {
			return global, nil
		}
		return "", fmt.Errorf(
			"script '%s' not found. Looked in:\n  local:  %s\n  global: %s",
			scriptName, local, global,
		)

	default: // ScopeGlobal
		if _, err := os.Stat(global); err == nil {
			return global, nil
		}
		return "", fmt.Errorf("script '%s' not found. Looked for it at: %s", scriptName, global)
	}
}

// Cmd represents the 'new' command
var Cmd = &cobra.Command{
	Use:   "new [script_name] [args...]",
	Short: "Run a Boiler command script (.bl)",
	Long: `Run a .bl command script to generate, inject, or modify code in your project.

Boiler automatically parses any flags (like --ts or --port=3000) and maps them to script variables.

Script resolution is scope-aware:
  - No boiler.local.json (or scope=global) : looks in ~/.boiler/commands/
  - scope=local in boiler.local.json        : looks in ./bl/commands/ first, falls back to ~/.boiler/commands/
  - --global flag                           : forces ~/.boiler/commands/ only
  - --local flag                            : forces ./bl/commands/ only`,
	Example: `  # Run the routes.bl script
  bl new routes

  # Run with positional arguments
  bl new routes user auth

  # Run with flags
  bl new routes --ts --port=3000

  # Force global commands directory
  bl new routes --global`,
	DisableFlagParsing: true, // Allows dynamic user-defined flags like --ts
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide a script name (e.g., 'bl new routes')")
		}

		if args[0] == "-h" || args[0] == "--help" {
			return cmd.Help()
		}

		scriptName := args[0]

		// Derive project root (same walk-up logic as before)
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		projectRoot := cwd
		if nearestConfig, cfgErr := config.FindNearestConfig(cwd); cfgErr == nil && nearestConfig != "" {
			projectRoot = filepath.Dir(nearestConfig)
		}

		hasGlobal := false
		hasLocal := false
		for _, arg := range os.Args {
			if arg == "--global" {
				hasGlobal = true
			} else if arg == "--local" {
				hasLocal = true
			}
		}
		for _, arg := range args {
			if arg == "--global" {
				hasGlobal = true
			} else if arg == "--local" {
				hasLocal = true
			}
		}

		scope := config.ScopeGlobal
		if config.Ctx != nil {
			scope = config.Ctx.Scope
		}

		if hasGlobal {
			scope = config.ScopeGlobal
		} else if hasLocal {
			scope = config.ScopeLocal
		}

		localCommandsDir := filepath.Join(projectRoot, constants.LocalBoilerDirName, constants.CommandsDirName)

		// cfg.Paths.Commands is set to ~/.boiler/commands/ (global default) or
		// ./bl/commands/ (when scope=local) by root.go PersistentPreRunE.
		// We always use the raw global path for fallback, so derive it independently.
		globalCommandsDir := cfg.Paths.Commands
		if config.Ctx != nil && config.Ctx.Scope == config.ScopeLocal {
			// Runtime path was overridden to local; recover the global path from defaults.
			globalCommandsDir = filepath.Join(config.DefaultConfig().Paths.Commands)
		}

		scriptPath, err := resolveScriptPath(scriptName, localCommandsDir, globalCommandsDir, scope, hasLocal)
		if err != nil {
			return err
		}

		// Build vars map from positional args and flags.
		// --global / --local / --verbose are consumed but not forwarded to the script.
		vars := make(map[string]string)
		positionalIndex := 1

		// Seed from BOILER_VAR_* env vars; CLI args below will override.
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

			if slices.Contains(constants.GlobalFlags, arg) {
				continue
			}

			if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
				flagRaw := strings.TrimLeft(arg, "-")
				if strings.Contains(flagRaw, "=") {
					parts := strings.SplitN(flagRaw, "=", 2)
					vars["bl__"+parts[0]] = parts[1]
				} else {
					vars["bl__"+flagRaw] = "true"
				}
			} else {
				vars[fmt.Sprintf("bl__%d", positionalIndex)] = arg
				positionalIndex++
			}
		}

		if logger != nil {
			logger.Info(fmt.Sprintf("Executing command script: %s", scriptPath))
		}

		// Switch to project root so relative paths inside the script resolve correctly.
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

		if err := engine.ParseAndExecute(scriptPath, vars); err != nil {
			return fmt.Errorf("script execution failed: %w", err)
		}

		if logger != nil {
			logger.Info(fmt.Sprintf("Successfully executed '%s'", scriptName))
		}
		return nil
	},
}
