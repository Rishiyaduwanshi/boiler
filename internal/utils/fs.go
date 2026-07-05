package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}
	if err := os.Chmod(dst, sourceInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

func CopyDir(src, dst string, ignorePatterns []string) error {
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	if err := os.MkdirAll(dst, sourceInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if ShouldIgnore(entry.Name(), ignorePatterns) {
			continue
		}

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath, ignorePatterns); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// ShouldIgnore reports whether name matches any of the given patterns.
// Supports both exact names and glob patterns (e.g. *.log).
func ShouldIgnore(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if name == pattern {
			return true
		}
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// CopyFileWithVariables copies a file and replaces template variables
// Variables format: bl__VAR_NAME will be replaced with provided values
// Also removes all metadata comments (__author, __desc, __var lines)
func CopyFileWithVariables(src, dst string, varReplacements map[string]string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Regex to match metadata lines
	metadataRe := regexp.MustCompile(`__(?:author|desc|version|var)\b`)

	scanner := bufio.NewScanner(sourceFile)
	writer := bufio.NewWriter(destFile)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip metadata comment lines
		if metadataRe.MatchString(line) {
			continue
		}

		// Replace variables (bl__VAR_NAME -> actual value)
		for varName, varValue := range varReplacements {
			line = strings.ReplaceAll(line, varName, varValue)
		}

		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}
	if err := os.Chmod(dst, sourceInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}
