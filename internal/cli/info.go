package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rishiyaduwanshi/boiler/internal/store"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [resource]",
	Short: "Show detailed information about a resource",
	Long: `Display detailed information about a stored snippet or stack.

Shows:
  - Full path in store
  - File size (for snippets)
  - File count and total size (for stacks)
  - Last modified time
  - Version information`,
	Example: `  # Show snippet info
  bl info errorHandler@1.js

  # Show stack info
  bl info express-api@1

  # Without version (shows all versions)
  bl info logger`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resource := args[0]
		logger.Info(fmt.Sprintf("Getting info for: %s", resource))

		if err := showInfo(resource); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func showInfo(resource string) error {
	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}

	fullName := utils.ParseResourceName(resource)

	if store.IsSnippet(resource) {
		return showSnippetInfo(st, fullName)
	}
	return showStackInfo(st, fullName)
}

func showSnippetInfo(st *store.Store, name string) error {
	path, ok := st.GetSnippet(name)
	if !ok {
		return fmt.Errorf(utils.ErrResourceNotFound, "snippet", name)
	}

	if !utils.FileExists(path) {
		return fmt.Errorf(utils.ErrResourceNotFound, "snippet file", path)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fmt.Printf("📄 Snippet: %s\n", name)
	fmt.Printf("   Path:     %s\n", path)
	fmt.Printf("   Size:     %d bytes\n", info.Size())
	fmt.Printf("   Modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
	fmt.Printf("   Type:     %s\n", filepath.Ext(path))

	return nil
}

func showStackInfo(st *store.Store, name string) error {
	path, ok := st.GetStack(name)
	if !ok {
		return fmt.Errorf(utils.ErrResourceNotFound, "stack", name)
	}

	if !utils.IsDirectory(path) {
		return fmt.Errorf(utils.ErrResourceNotFound, "stack directory", path)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to get directory info: %w", err)
	}

	// Count files
	fileCount := 0
	dirCount := 0
	if walkErr := filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			dirCount++
		} else {
			fileCount++
		}
		return nil
	}); walkErr != nil {
		return fmt.Errorf("failed to read stack contents: %w", walkErr)
	}

	fmt.Printf("📦 Stack: %s\n", name)
	fmt.Printf("   Path:        %s\n", path)
	fmt.Printf("   Files:       %d\n", fileCount)
	fmt.Printf("   Directories: %d\n", dirCount-1) // -1 to exclude root
	fmt.Printf("   Modified:    %s\n", info.ModTime().Format("2006-01-02 15:04:05"))

	return nil
}
