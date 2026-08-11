package add

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/store"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

func AddSnippet(st *store.Store, name, destPath string, opts Options, cfg *config.Config, logger *utils.Logger) error {
	snippetPath, ok := st.GetSnippet(name)
	if !ok {
		return fmt.Errorf(utils.ErrResourceNotFound, "snippet", name)
	}

	if !utils.FileExists(snippetPath) {
		return fmt.Errorf(utils.ErrResourceNotFound, "snippet file", snippetPath)
	}

	// Extract base name without version: errorHandler@1.js -> errorHandler.js
	baseName, _, ext := store.ParseResourceName(name)
	destFileName := baseName + ext
	if opts.Name != "" {
		destFileName = opts.Name
	}
	destFile := filepath.Join(destPath, destFileName)

	if utils.FileExists(destFile) && !opts.Force {
		return fmt.Errorf(utils.ErrFileAlreadyExists, destFile)
	}

	if err := ProcessSnippetFile(snippetPath, destFile, cfg); err != nil {
		return err
	}

	fmt.Printf(utils.MsgSnippetAdded, name, destFile)
	if logger != nil {
		logger.Info(fmt.Sprintf("Snippet added: %s -> %s", name, destFile))
	}
	return nil
}

// ProcessSnippetFile parses metadata, prompts for variables, and copies the content to destFile.
func ProcessSnippetFile(snippetPath, destFile string, cfg *config.Config) error {
	// Parse snippet metadata to check for variables
	meta, err := utils.ParseSnippetMetadata(snippetPath)
	if err != nil {
		return fmt.Errorf("failed to parse snippet metadata: %w", err)
	}

	// Prompt user for variable values if variables exist
	varReplacements := make(map[string]string)
	if len(meta.Variables) > 0 {
		fmt.Println("Template variables found:")
		for varName, defaultValue := range meta.Variables {
			// If value is provided via Env or Config, skip the prompt and use it silently
			if val, ok, _ := utils.LookupVar(cfg.Vars, varName); ok && val != "" {
				varReplacements[varName] = val
				continue
			}

			defaultValue = utils.ResolveSnippetVarDefault(varName, defaultValue, cfg.Vars)
			prompt := fmt.Sprintf("  %s", varName)
			value, err := utils.PromptWithDefault(prompt, defaultValue)
			if err != nil {
				return fmt.Errorf("failed to read variable input: %w", err)
			}
			varReplacements[varName] = value
		}
	}

	// Copy file with variable replacement
	if err := utils.CopyFileWithVariables(snippetPath, destFile, varReplacements); err != nil {
		return fmt.Errorf("failed to copy snippet: %w", err)
	}

	return nil
}

func AddStack(st *store.Store, name, destPath string, opts Options, logger *utils.Logger) error {
	stackPath, ok := st.GetStack(name)
	if !ok {
		return fmt.Errorf(utils.ErrResourceNotFound, "stack", name)
	}

	if !utils.IsDirectory(stackPath) {
		return fmt.Errorf(utils.ErrResourceNotFound, "stack directory", stackPath)
	}

	finalDestPath, err := copyStackToDestination(stackPath, name, destPath, opts)
	if err != nil {
		return err
	}

	fmt.Printf(utils.MsgStackAdded, name, finalDestPath)
	if logger != nil {
		logger.Info(fmt.Sprintf("Stack added: %s -> %s", name, finalDestPath))
	}
	return nil
}

// validateStackDestination reports whether a non-spread stack placement would
// conflict with an existing path. Call this before any remote download so
// boiler use / bl add --remote fail fast without network I/O (issue #61).
//
// Spread mode still needs stack contents for per-file conflict checks, so this
// only guards the non-spread dest/stackDir path (and a non-directory dest when
// --spread is set).
func validateStackDestination(stackName, destPath string, opts Options) error {
	if opts.Force {
		return nil
	}

	if opts.Spread {
		if utils.FileExists(destPath) && !utils.IsDirectory(destPath) {
			return fmt.Errorf("destination '%s' must be a directory", destPath)
		}
		return nil
	}

	stackDir := stackDirectoryName(stackName)
	if opts.Name != "" {
		stackDir = opts.Name
	}
	finalDestPath := filepath.Join(destPath, stackDir)
	if utils.FileExists(finalDestPath) {
		return fmt.Errorf(utils.ErrDestAlreadyExists, finalDestPath)
	}
	return nil
}

func copyStackToDestination(stackPath, stackName, destPath string, opts Options) (string, error) {
	ignorePatterns := utils.DefaultIgnorePatterns

	if opts.Spread {
		if err := validateSpreadDestination(stackPath, destPath, ignorePatterns, opts); err != nil {
			return "", err
		}
		if err := utils.CopyDir(stackPath, destPath, ignorePatterns); err != nil {
			return "", fmt.Errorf("failed to copy stack: %w", err)
		}
		return destPath, nil
	}

	if err := validateStackDestination(stackName, destPath, opts); err != nil {
		return "", err
	}

	stackDir := stackDirectoryName(stackName)
	if opts.Name != "" {
		stackDir = opts.Name
	}
	finalDestPath := filepath.Join(destPath, stackDir)

	if err := utils.CopyDir(stackPath, finalDestPath, ignorePatterns); err != nil {
		return "", fmt.Errorf("failed to copy stack: %w", err)
	}

	return finalDestPath, nil
}

func validateSpreadDestination(stackPath, destPath string, ignorePatterns []string, opts Options) error {
	if opts.Force {
		return nil
	}

	if !utils.FileExists(destPath) {
		return nil
	}

	if !utils.IsDirectory(destPath) {
		return fmt.Errorf("destination '%s' must be a directory", destPath)
	}

	entries, err := os.ReadDir(stackPath)
	if err != nil {
		return fmt.Errorf("failed to inspect stack contents: %w", err)
	}

	for _, entry := range entries {
		if utils.ShouldIgnore(entry.Name(), ignorePatterns) {
			continue
		}

		target := filepath.Join(destPath, entry.Name())
		if utils.FileExists(target) {
			return fmt.Errorf(utils.ErrDestAlreadyExists, target)
		}
	}

	return nil
}

func stackDirectoryName(resourceName string) string {
	baseName, _, _ := store.ParseResourceName(resourceName)
	if baseName == "" {
		return resourceName
	}
	return baseName
}
