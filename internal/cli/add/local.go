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
			defaultValue = utils.ResolveSnippetVarDefault(varName, defaultValue, cfg.Vars)
			prompt := fmt.Sprintf("  %s", varName)
			value, err := utils.PromptWithDefault(prompt, defaultValue)
			if err != nil {
				return fmt.Errorf("failed to read variable input: %w", err)
			}
			varReplacements[varName] = value
		}
	}

	// Extract base name without version: errorHandler@1.js -> errorHandler.js
	baseName, _, ext := store.ParseResourceName(name)
	destFileName := baseName + ext
	destFile := filepath.Join(destPath, destFileName)

	if utils.FileExists(destFile) && !opts.Force {
		return fmt.Errorf(utils.ErrFileAlreadyExists, destFile)
	}

	// Copy file with variable replacement
	if err := utils.CopyFileWithVariables(snippetPath, destFile, varReplacements); err != nil {
		return fmt.Errorf("failed to copy snippet: %w", err)
	}

	fmt.Printf(utils.MsgSnippetAdded, name, destFile)
	if logger != nil {
		logger.Info(fmt.Sprintf("Snippet added: %s -> %s", name, destFile))
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

	stackDir := stackDirectoryName(stackName)
	finalDestPath := filepath.Join(destPath, stackDir)
	if utils.FileExists(finalDestPath) && !opts.Force {
		return "", fmt.Errorf(utils.ErrDestAlreadyExists, finalDestPath)
	}

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
