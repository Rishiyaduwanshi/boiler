package config

import (
	"os"
	"path/filepath"

	"github.com/rishiyaduwanshi/boiler/internal/constants"
)

// FindNearestConfig walks up the directory tree from cwd until it finds a `.boiler/config.json`
// Returns the path if found, or an empty string if not found.
func FindNearestConfig(cwd string) (string, error) {
	currentDir := cwd

	for {
		configPath := filepath.Join(currentDir, constants.LocalConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		// Get the parent directory
		parentDir := filepath.Dir(currentDir)

		// Stop if we've reached the root directory (parent is same as current)
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}

	return "", nil
}

// GlobalConfigPath returns the path to the global config file (~/.boiler/boiler.conf.json).
func GlobalConfigPath() (string, error) {
	rootPath := getRootPath()
	return filepath.Join(rootPath, constants.GlobalConfigFileName), nil
}

// DefaultLocalConfigPath returns the path where a local config SHOULD be created
// if the user is in a project that doesn't have one yet.
func DefaultLocalConfigPath(cwd string) string {
	return filepath.Join(cwd, constants.LocalConfigFileName)
}
