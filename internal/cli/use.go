package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rishiyaduwanshi/boiler/internal/remote"
	"github.com/rishiyaduwanshi/boiler/internal/store"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use [resource]",
	Short: "Fetch a remote resource directly without saving to local store",
	Long: `Fetch a snippet or stack from any remote source and place it in the
current directory - no local store involved, no registry lookup needed.

Unlike 'bl add -r', this command is purely one-shot:
  - Downloads the resource
  - Copies it to the destination
  - Does NOT save it to your local store

Provider is auto-detected from the URL (GitHub, GitLab, Bitbucket, generic).
Both .zip and .tar.gz archives are supported and auto-detected from the URL.

Supported formats:
  owner/repo                             GitHub repo as stack (default branch)
  owner/repo:path/to/file.js            File from GitHub repo
  https://github.com/owner/repo         GitHub full URL
  https://gitlab.com/owner/repo         GitLab repo
  https://bitbucket.org/owner/repo      Bitbucket repo
  https://anysite.com/stack.zip         Direct zip archive
  https://anysite.com/stack.tar.gz      Direct tar.gz archive
  https://anysite.com/file.js           Direct file (snippet)`,
	Example: `  # GitHub repo as stack
  bl use alice/my-express-stack

  # GitLab repo
  bl use https://gitlab.com/alice/my-stack

  # Bitbucket repo
  bl use https://bitbucket.org/alice/my-stack

  # File from GitHub repo (snippet)
  bl use alice/snippets:js/errorHandler.js

  # Direct zip archive (any site)
  bl use https://mysite.com/templates/express.zip

  # Direct tar.gz archive
  bl use https://mysite.com/stack.tar.gz

  # Direct file URL (snippet)
  bl use https://mysite.com/snippets/logger.js

	# Resource from config variable
	bl use :starter_stack

  # Into a specific folder
  bl use alice/my-stack --to ./new-project`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remotePath, err := utils.ResolveInputToken(args[0], "resource", cfg.Vars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		destPath, err := utils.ResolveInputToken(useTo, "destination", cfg.Vars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := useResource(remotePath, destPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var useTo string

func init() {
	useCmd.Flags().StringVarP(&useTo, "to", "t", ".", "Destination path")
	rootCmd.AddCommand(useCmd)
}

func useResource(remotePath, destPath string) error {
	if destPath == "" {
		destPath = "."
	}

	owner, _, subPath := store.ParseRemotePath(remotePath)

	// Decide: snippet (subPath has a file extension) or stack
	isSnippet := filepath.Ext(subPath) != ""

	// Direct URL to a file (not an archive) - treat as snippet
	if !isSnippet && (isDirectFileURL(remotePath)) {
		isSnippet = true
	}

	if isSnippet {
		fileName := filepath.Base(subPath)
		if fileName == "" || fileName == "." {
			fileName = filepath.Base(remotePath)
		}
		destFile := filepath.Join(destPath, fileName)

		if err := remote.FetchSnippet(remotePath, destFile); err != nil {
			return err
		}
		fmt.Printf("✓ %s → %s\n", fileName, destFile)
		return nil
	}

	// Stack: owner/repo, full URL, or direct archive URL
	_ = owner // used inside FetchStack via remotePath
	if err := remote.FetchStack(remotePath, destPath); err != nil {
		return err
	}
	return nil
}

// isDirectFileURL returns true when the URL points to a plain file (not an archive).
func isDirectFileURL(url string) bool {
	ext := filepath.Ext(url)
	switch ext {
	case ".zip", ".gz", ".tgz", "": // archive or no ext → not a plain file
		return false
	default:
		return true
	}
}
