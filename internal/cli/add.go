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

var addCmd = &cobra.Command{
	Use:   "add [resource] [destination]",
	Short: "Add a snippet or stack to boiler/ by default",
	Long: `Add a stored snippet or stack to ./boiler by default.

Destination:
	Use optional [destination] to override the default path:
		bl add logger .
		bl add logger ./src/utils
		bl add logger /absolute/path

Stack placement:
	By default, stacks are copied inside a stack-named folder:
		bl add express@1            -> ./boiler/express
	Use --spread to copy stack contents directly into destination:
		bl add express@1 --spread   -> contents in ./boiler

The command copies resources from your store. For snippets with a single version,
you can use just the name (e.g., 'errorHandler' will auto-select version 1).
For multiple versions, you'll be prompted to choose.

Template Variables:
  Snippets can contain template variables using the format: bl__VAR_NAME
  When adding a snippet with variables, you'll be prompted to provide values:
    - Default values are shown in brackets (from __var declarations)
    - Press Enter to use default or type a custom value
    - Variables are replaced and metadata comments are removed in the final file

Stacks are also versioned and can be added by name or with explicit version.

Remote Resources:
  Use -r flag to fetch from remote source and save to local store.
  Provider is auto-detected from the URL (GitHub, GitLab, Bitbucket, generic).
  Resource is cached locally - subsequent uses don't need -r.

  For one-shot fetch without saving to local store, use 'bl use' instead.

    1. Registry:           bl add express@1 -r
       (registry set via: bl conf --set-registry <url>)

    2. GitHub short:       bl add owner/repo -r
    3. GitHub full URL:    bl add https://github.com/owner/repo -r
    4. GitLab:             bl add https://gitlab.com/owner/repo -r
    5. Bitbucket:          bl add https://bitbucket.org/owner/repo -r

    6. File from repo:     bl add owner/repo:path/to/file.js -r
    7. Direct file URL:    bl add https://site.com/file.js -r
    8. Direct archive:     bl add https://site.com/stack.zip -r
    9. Custom domain file: bl add site.com:path/file.js -r

   10. One-time registry:  bl add express@1 -r --registry https://github.com/other/boiler`,
	Example: `  # Add snippet (auto-detects if single version)
  bl add errorHandler

	# Add to current directory
	bl add errorHandler .

	# Add to a custom destination
	bl add errorHandler ./src/utils

  # Add snippet with template variables
  bl add apiClient
  # Prompts: bl__API_URL [http://localhost:3000]: https://api.example.com
  #          bl__API_KEY [your-key]: abc123xyz
  # Output: Clean file with variables replaced, no metadata comments

  # Add specific version
  bl add logger@2.js

	# Add stack into boiler/express-api
  bl add express-api@1

	# Add stack contents directly into destination
	bl add express-api@1 --spread

  # Force overwrite
  bl add middleware --force

  # Remote: from configured registry
  bl add express@1 -r

  # Remote: GitHub short format
  bl add rishiyaduwanshi/boiler-express -r

  # Remote: GitLab
  bl add https://gitlab.com/alice/my-stack -r

  # Remote: Bitbucket
  bl add https://bitbucket.org/alice/my-stack -r

  # Remote: file inside GitHub repo
  bl add rishiyaduwanshi/boiler-snippets:js/errorHandler.js -r

  # Remote: direct file URL
  bl add https://mysite.com/snippets/helper.js -r

  # Remote: direct archive URL
  bl add https://mysite.com/stack.zip -r

  # Remote: one-time registry override
  bl add express@1 -r --registry https://github.com/myorg/boiler

  # One-shot fetch without saving to store (no -r needed)
  bl use alice/my-stack`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		resource := args[0]
		positionalDest := ""
		if len(args) == 2 {
			positionalDest = args[1]
		}

		destPath := resolveAddDestination(positionalDest)
		logger.Info(fmt.Sprintf("Adding resource: %s -> %s", resource, destPath))

		if err := addResource(resource, destPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

const addDefaultDestination = "boiler"

func resolveAddDestination(positionalDest string) string {
	if positionalDest == "" {
		return addDefaultDestination
	}
	return filepath.Clean(positionalDest)
}

func addResource(resource, destPath string) error {
	if addRemote {
		return addResourceFromRemote(resource, destPath)
	}

	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}

	baseName, version, ext := store.ParseResourceName(resource)

	// Snippet (has file extension)
	if ext != "" {
		if addSpread {
			return fmt.Errorf("--spread is only supported for stacks")
		}
		if version != "" {
			return addSnippet(st, baseName+"@"+version+ext, destPath)
		}
		matches := utils.FindMatchingResources(st.ListSnippets(), baseName, ext)
		if len(matches) == 0 {
			return fmt.Errorf(utils.ErrResourceNotFound, "snippet", resource)
		}
		selected, err := utils.PickFromList(baseName+ext, matches)
		if err != nil {
			return err
		}
		return addSnippet(st, selected, destPath)
	}

	// No extension - try as stack first
	if _, ok := st.GetStack(resource); ok {
		return addStack(st, resource, destPath)
	}

	// Try stack lookup by base name
	stackMatches := utils.FindMatchingResources(st.ListStacks(), baseName, "")
	if len(stackMatches) > 0 {
		selected, err := utils.PickFromList(baseName, stackMatches)
		if err != nil {
			return err
		}
		return addStack(st, selected, destPath)
	}

	// Fall back to snippet lookup (name without version/ext)
	matches := utils.FindMatchingResources(st.ListSnippets(), baseName, "")
	if len(matches) == 0 {
		return fmt.Errorf(utils.ErrResourceNotFound, "stack or snippet", resource)
	}
	selected, err := utils.PickFromList(baseName, matches)
	if err != nil {
		return err
	}
	return addSnippet(st, selected, destPath)
}

func addSnippet(st *store.Store, name, destPath string) error {
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

	if utils.FileExists(destFile) && !addForce {
		return fmt.Errorf(utils.ErrFileAlreadyExists, destFile)
	}

	// Copy file with variable replacement
	if err := utils.CopyFileWithVariables(snippetPath, destFile, varReplacements); err != nil {
		return fmt.Errorf("failed to copy snippet: %w", err)
	}

	fmt.Printf(utils.MsgSnippetAdded, name, destFile)
	logger.Info(fmt.Sprintf("Snippet added: %s -> %s", name, destFile))
	return nil
}

func addStack(st *store.Store, name, destPath string) error {
	stackPath, ok := st.GetStack(name)
	if !ok {
		return fmt.Errorf(utils.ErrResourceNotFound, "stack", name)
	}

	if !utils.IsDirectory(stackPath) {
		return fmt.Errorf(utils.ErrResourceNotFound, "stack directory", stackPath)
	}

	finalDestPath, err := copyStackToDestination(stackPath, name, destPath)
	if err != nil {
		return err
	}

	fmt.Printf(utils.MsgStackAdded, name, finalDestPath)
	logger.Info(fmt.Sprintf("Stack added: %s -> %s", name, finalDestPath))
	return nil
}

func copyStackToDestination(stackPath, stackName, destPath string) (string, error) {
	ignorePatterns := utils.DefaultIgnorePatterns

	if addSpread {
		if err := validateSpreadDestination(stackPath, destPath, ignorePatterns); err != nil {
			return "", err
		}
		if err := utils.CopyDir(stackPath, destPath, ignorePatterns); err != nil {
			return "", fmt.Errorf("failed to copy stack: %w", err)
		}
		return destPath, nil
	}

	stackDir := stackDirectoryName(stackName)
	finalDestPath := filepath.Join(destPath, stackDir)
	if utils.FileExists(finalDestPath) && !addForce {
		return "", fmt.Errorf(utils.ErrDestAlreadyExists, finalDestPath)
	}

	if err := utils.CopyDir(stackPath, finalDestPath, ignorePatterns); err != nil {
		return "", fmt.Errorf("failed to copy stack: %w", err)
	}

	return finalDestPath, nil
}

func validateSpreadDestination(stackPath, destPath string, ignorePatterns []string) error {
	if addForce {
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
		if isIgnoredEntry(entry.Name(), ignorePatterns) {
			continue
		}

		target := filepath.Join(destPath, entry.Name())
		if utils.FileExists(target) {
			return fmt.Errorf(utils.ErrDestAlreadyExists, target)
		}
	}

	return nil
}

func isIgnoredEntry(name string, patterns []string) bool {
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

func stackDirectoryName(resourceName string) string {
	baseName, _, _ := store.ParseResourceName(resourceName)
	if baseName == "" {
		return resourceName
	}
	return baseName
}

// addResourceFromRemote fetches and adds a resource from remote registry
func addResourceFromRemote(resource, destPath string) error {
	// Check if resource is direct GitHub path (owner/repo format)
	if store.IsRemotePath(resource) {
		return addDirectRemoteResource(resource, destPath)
	}

	// Use custom registry if provided, otherwise use config
	registryURL := cfg.Registry
	if addRegistry != "" {
		registryURL = addRegistry
	}

	// Initialize remote store
	remoteStoreHandler, err := remote.NewRemoteStore(registryURL)
	if err != nil {
		return fmt.Errorf("failed to initialize remote store: %w", err)
	}

	// Load remote metadata (boiler.meta.json from GitHub)
	fmt.Println("🔄 Fetching remote registry...")
	remoteStore, err := remoteStoreHandler.LoadFromURL()
	if err != nil {
		return fmt.Errorf("failed to load remote registry: %w", err)
	}

	// Parse resource name
	baseName, _, ext := store.ParseResourceName(resource)

	// Try as snippet first
	if ext != "" {
		if addSpread {
			return fmt.Errorf("--spread is only supported for stacks")
		}
		remotePath, exists := remoteStore.GetSnippet(resource)
		if !exists {
			matches := utils.FindMatchingResources(remoteStore.ListSnippets(), baseName, ext)
			if len(matches) == 0 {
				return fmt.Errorf("snippet '%s' not found in remote registry", resource)
			}
			selected, err := utils.PickFromList(baseName+ext, matches)
			if err != nil {
				return err
			}
			resource = selected
			remotePath, exists = remoteStore.GetSnippet(resource)
		}

		// Check if it's a remote path
		if !store.IsRemotePath(remotePath) {
			return fmt.Errorf("snippet '%s' does not have a valid remote location", resource)
		}

		// Download snippet to local store first
		localStorePath := filepath.Join(cfg.Paths.Snippets, filepath.Dir(remotePath))
		if err := os.MkdirAll(localStorePath, 0755); err != nil {
			return fmt.Errorf("failed to create local store directory: %w", err)
		}

		// Parse remote path to get filename
		_, _, remotefile := store.ParseRemotePath(remotePath)
		localDestPath := filepath.Join(cfg.Paths.Snippets, filepath.Base(remotefile))

		// Fetch snippet
		if err := remote.FetchSnippet(remotePath, localDestPath); err != nil {
			return err
		}

		// Add to local store metadata
		st, err := utils.LoadStore(cfg.Paths.Store)
		if err != nil {
			return err
		}
		if err := st.AddSnippet(resource, localDestPath); err != nil {
			return err
		}

		// Copy to destination
		selectedBaseName, _, selectedExt := store.ParseResourceName(resource)
		destFileName := selectedBaseName + selectedExt
		finalDestPath := filepath.Join(destPath, destFileName)
		if err := utils.CopyFileWithVariables(localDestPath, finalDestPath, nil); err != nil {
			return fmt.Errorf("failed to copy snippet: %w", err)
		}

		fmt.Printf(utils.MsgSnippetAdded, resource, finalDestPath)
		logger.Info(fmt.Sprintf("Remote snippet added: %s -> %s", resource, finalDestPath))
		return nil
	}

	// Try as stack
	remotePath, exists := remoteStore.GetStack(resource)
	if !exists {
		matches := utils.FindMatchingResources(remoteStore.ListStacks(), baseName, "")
		if len(matches) == 0 {
			return fmt.Errorf("resource '%s' not found in remote registry", resource)
		}
		selected, err := utils.PickFromList(baseName, matches)
		if err != nil {
			return err
		}
		resource = selected
		remotePath, exists = remoteStore.GetStack(resource)
	}

	// Check if it's a remote path
	if !store.IsRemotePath(remotePath) {
		return fmt.Errorf("stack '%s' does not have a valid remote location", resource)
	}

	// Download stack to local store first
	localStackPath := filepath.Join(cfg.Paths.Stacks, stackDirectoryName(resource))

	// Fetch stack
	if err := remote.FetchStack(remotePath, localStackPath); err != nil {
		return err
	}

	// Add to local store metadata
	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}
	if err := st.AddStack(resource, localStackPath); err != nil {
		return err
	}

	finalDestPath, err := copyStackToDestination(localStackPath, resource, destPath)
	if err != nil {
		return err
	}

	fmt.Printf(utils.MsgStackAdded, resource, finalDestPath)
	logger.Info(fmt.Sprintf("Remote stack added: %s -> %s", resource, finalDestPath))
	return nil
}

// addDirectRemoteResource adds a resource directly from GitHub without registry lookup
// Supports formats: owner/repo, owner/repo:path, owner/repo@branch:path
func addDirectRemoteResource(remotePath, destPath string) error {
	// Parse remote path
	owner, repo, subPath := store.ParseRemotePath(remotePath)
	if owner == "" || repo == "" {
		return fmt.Errorf("invalid remote path format: %s (expected: owner/repo or owner/repo:path)", remotePath)
	}

	fmt.Printf("📥 Fetching directly from %s/%s...\n", owner, repo)

	// Determine if it's a stack (no extension in subPath) or snippet (has extension)
	isSnippet := filepath.Ext(subPath) != ""

	if isSnippet {
		if addSpread {
			return fmt.Errorf("--spread is only supported for stacks")
		}

		// Download snippet
		localStorePath := filepath.Join(cfg.Paths.Snippets, filepath.Dir(subPath))
		if err := os.MkdirAll(localStorePath, 0755); err != nil {
			return fmt.Errorf("failed to create local store directory: %w", err)
		}

		localDestPath := filepath.Join(cfg.Paths.Snippets, filepath.Base(subPath))

		// Fetch snippet
		if err := remote.FetchSnippet(remotePath, localDestPath); err != nil {
			return err
		}

		// Generate resource name: filename without path
		resourceName := filepath.Base(subPath)

		// Add to local store metadata
		st, err := utils.LoadStore(cfg.Paths.Store)
		if err != nil {
			return err
		}
		if err := st.AddSnippet(resourceName, localDestPath); err != nil {
			return err
		}

		// Copy to destination
		finalDestPath := filepath.Join(destPath, resourceName)
		if err := utils.CopyFileWithVariables(localDestPath, finalDestPath, nil); err != nil {
			return fmt.Errorf("failed to copy snippet: %w", err)
		}

		fmt.Printf(utils.MsgSnippetAdded, resourceName, finalDestPath)
		logger.Info(fmt.Sprintf("Direct remote snippet added: %s -> %s", remotePath, finalDestPath))
		return nil
	}

	// It's a stack
	localStackPath := filepath.Join(cfg.Paths.Stacks, repo)

	// Fetch stack
	if err := remote.FetchStack(remotePath, localStackPath); err != nil {
		return err
	}

	// Generate resource name: repo name
	resourceName := repo

	// Add to local store metadata
	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}
	if err := st.AddStack(resourceName, localStackPath); err != nil {
		return err
	}

	finalDestPath, err := copyStackToDestination(localStackPath, resourceName, destPath)
	if err != nil {
		return err
	}

	fmt.Printf(utils.MsgStackAdded, resourceName, finalDestPath)
	logger.Info(fmt.Sprintf("Direct remote stack added: %s -> %s", remotePath, finalDestPath))
	return nil
}

var (
	addRemote   bool
	addSpread   bool
	addForce    bool
	addRegistry string
)

func init() {
	addCmd.Flags().BoolVarP(&addRemote, "remote", "r", false, "Fetch from remote registry")
	addCmd.Flags().BoolVar(&addSpread, "spread", false, "Spread stack contents directly into destination")
	addCmd.Flags().BoolVarP(&addForce, FlagForce, FlagForceShort, false, DescForce)
	addCmd.Flags().StringVar(&addRegistry, "registry", "", "Custom registry URL (overrides config)")
}
