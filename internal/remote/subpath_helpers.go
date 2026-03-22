package remote

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

func normalizeRemoteSubPath(subPath string) string {
	trimmed := strings.TrimSpace(subPath)
	trimmed = strings.TrimPrefix(trimmed, "./")
	trimmed = strings.TrimPrefix(trimmed, "/")
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" {
		return "."
	}
	return trimmed
}

func prepareSubtreeDestination(destPath string) error {
	if err := os.RemoveAll(destPath); err != nil {
		return fmt.Errorf("failed to reset destination: %w", err)
	}
	if err := utils.EnsureDir(destPath); err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	return nil
}

func writeSubtreeFile(basePath, absolutePath, destRoot string, data []byte) error {
	relPath := strings.TrimPrefix(absolutePath, basePath+"/")
	if relPath == absolutePath {
		relPath = filepath.Base(absolutePath)
	}

	destFile := filepath.Join(destRoot, filepath.FromSlash(relPath))
	if err := utils.EnsureDir(filepath.Dir(destFile)); err != nil {
		return fmt.Errorf("failed to create destination for %s: %w", relPath, err)
	}

	if err := os.WriteFile(destFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", relPath, err)
	}

	return nil
}