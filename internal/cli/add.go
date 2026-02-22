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
	Use:   "add [resource]",
	Short: "Add a snippet or stack to current directory",
	Long: `Add a stored snippet or stack to your current directory.

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
  Use -r flag to fetch from remote registry or directly from GitHub/URLs:
    1. From registry: bl add express@1 -r
       (Uses registry configured in config or --registry flag)
    
    2. Direct from GitHub: bl add owner/repo -r
       (Fetches entire repo as stack)
    
    3. Direct snippet: bl add owner/repo:path/to/file.js -r
       (Fetches single file)
    
    4. Direct URL: bl add https://yourdomain.com/path/file.js -r
       (Downloads from any URL)
    
    5. Custom domain: bl add yourdomain.com:path/file.js -r
       (Assumes HTTPS)
    
    6. Custom registry: bl add express@1 -r --registry https://github.com/other/boiler
       (One-time registry override)`,
	Example: `  # Add snippet (auto-detects if single version)
  bl add errorHandler

  # Add snippet with template variables
  bl add apiClient
  # Prompts: bl__API_URL [http://localhost:3000]: https://api.example.com
  #          bl__API_KEY [your-key]: abc123xyz
  # Output: Clean file with variables replaced, no metadata comments

  # Add specific version
  bl add logger@2.js

  # Add to specific directory
  bl add config --to ./src/utils

  # Add stack
  bl add express-api@1

  # Force overwrite
  bl add middleware --force

  # Remote: From registry
  bl add express@1 -r

  # Remote: Direct from GitHub (entire repo)
  bl add rishiyaduwanshi/boiler-express -r

  # Remote: Direct snippet from GitHub
  bl add rishiyaduwanshi/boiler-snippets:js/errorHandler.js -r

  # Remote: From custom website (direct URL)
  bl add https://iamabhinav.dev/snippets/helper.js -r

  # Remote: From custom domain
  bl add iamabhinav.dev:snippets/validator.js -r

  # Remote: Custom registry
  bl add express@1 -r --registry https://github.com/myorg/boiler`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resource := args[0]
		logger.Info(fmt.Sprintf("Adding resource: %s", resource))

		if err := addResource(resource); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func addResource(resource string) error {
	// Handle remote fetch
	if addRemote {
		return addResourceFromRemote(resource)
	}

	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}

	destPath := addTo
	if destPath == "" {
		destPath = "."
	}

	// Parse resource name to extract parts
	baseName, version, ext := store.ParseResourceName(resource)

	// If it's a snippet (has extension)
	if ext != "" {
		// If version is explicitly provided, use it directly
		if version != "" {
			fullName := baseName + "@" + version + ext
			return addSnippet(st, fullName, destPath)
		}

		// No version specified - find matching snippets by name and extension
		matchingSnippets := findMatchingSnippetsByNameAndExt(st, baseName, ext)
		
		if len(matchingSnippets) == 0 {
			return fmt.Errorf(utils.ErrResourceNotFound, "snippet", resource)
		}

		// If only one version exists, use it automatically
		if len(matchingSnippets) == 1 {
			return addSnippet(st, matchingSnippets[0], destPath)
		}

		// Multiple versions - prompt user to choose
		fmt.Printf("Multiple versions found for '%s%s':\n", baseName, ext)
		for i, name := range matchingSnippets {
			fmt.Printf("  %d. %s\n", i+1, name)
		}

		choice, err := utils.Prompt(fmt.Sprintf("Enter version number (1-%d): ", len(matchingSnippets)))
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		var selectedIdx int
		fmt.Sscanf(choice, "%d", &selectedIdx)
		if selectedIdx < 1 || selectedIdx > len(matchingSnippets) {
			return fmt.Errorf("invalid choice")
		}

		return addSnippet(st, matchingSnippets[selectedIdx-1], destPath)
	}

	// No extension - could be stack or snippet name without version/extension
	// First check if it exists as a stack
	_, stackExists := st.GetStack(resource)
	if stackExists {
		return addStack(st, resource, destPath)
	}

	// Not a stack, try to find matching snippets by base name only
	matchingSnippets := findMatchingSnippets(st, baseName)
	if len(matchingSnippets) == 0 {
		return fmt.Errorf(utils.ErrResourceNotFound, "stack or snippet", resource)
	}

	// If only one version exists, use it automatically
	if len(matchingSnippets) == 1 {
		return addSnippet(st, matchingSnippets[0], destPath)
	}

	// Multiple versions - prompt user to choose
	fmt.Printf("Multiple versions found for '%s':\n", baseName)
	for i, name := range matchingSnippets {
		fmt.Printf("  %d. %s\n", i+1, name)
	}

	choice, err := utils.Prompt(fmt.Sprintf("Enter version number (1-%d): ", len(matchingSnippets)))
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Parse choice
	var selectedIdx int
	fmt.Sscanf(choice, "%d", &selectedIdx)
	if selectedIdx < 1 || selectedIdx > len(matchingSnippets) {
		return fmt.Errorf("invalid choice")
	}

	return addSnippet(st, matchingSnippets[selectedIdx-1], destPath)
}

// findMatchingSnippets finds all snippets that match the given name (without version/extension)
func findMatchingSnippets(st *store.Store, name string) []string {
	allSnippets := st.ListSnippets()
	var matches []string

	for _, snippet := range allSnippets {
		snippetName, _, _ := store.ParseResourceName(snippet)
		if snippetName == name {
			matches = append(matches, snippet)
		}
	}

	return matches
}

// findMatchingSnippetsByNameAndExt finds all snippets matching both name and extension
func findMatchingSnippetsByNameAndExt(st *store.Store, name, ext string) []string {
	allSnippets := st.ListSnippets()
	var matches []string

	for _, snippet := range allSnippets {
		snippetName, _, snippetExt := store.ParseResourceName(snippet)
		if snippetName == name && snippetExt == ext {
			matches = append(matches, snippet)
		}
	}

	return matches
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

	if utils.FileExists(destPath) && destPath != "." && !addForce {
			return fmt.Errorf(utils.ErrDestAlreadyExists, destPath)
	}

	ignorePatterns := []string{"node_modules", ".git", ".DS_Store", "Thumbs.db"}
	if err := utils.CopyDir(stackPath, destPath, ignorePatterns); err != nil {
		return fmt.Errorf("failed to copy stack: %w", err)
	}

	fmt.Printf(utils.MsgStackAdded, name, destPath)
	logger.Info(fmt.Sprintf("Stack added: %s -> %s", name, destPath))
	return nil
}

// addResourceFromRemote fetches and adds a resource from remote registry
func addResourceFromRemote(resource string) error {
	// Check if resource is direct GitHub path (owner/repo format)
	if store.IsRemotePath(resource) {
		return addDirectRemoteResource(resource)
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

	destPath := addTo
	if destPath == "" {
		destPath = "."
	}

	// Parse resource name
	baseName, _, ext := store.ParseResourceName(resource)

	// Try as snippet first
	if ext != "" {
		remotePath, exists := remoteStore.GetSnippet(resource)
		if !exists {
			// Try finding matching snippets by base name
			matches := findMatchingRemoteSnippets(remoteStore, baseName, ext)
			if len(matches) == 0 {
				return fmt.Errorf("snippet '%s' not found in remote registry", resource)
			}

			// Auto-select if only one match
			if len(matches) == 1 {
				resource = matches[0]
				remotePath, exists = remoteStore.GetSnippet(resource)
			} else {
				// Multiple versions - prompt user
				fmt.Printf("Multiple versions found for '%s':\n", baseName)
				for i, name := range matches {
					fmt.Printf("  %d. %s\n", i+1, name)
				}

				choice, err := utils.Prompt(fmt.Sprintf("Enter version number (1-%d): ", len(matches)))
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}

				var selectedIdx int
				fmt.Sscanf(choice, "%d", &selectedIdx)
				if selectedIdx < 1 || selectedIdx > len(matches) {
					return fmt.Errorf("invalid choice")
				}

				resource = matches[selectedIdx-1]
				remotePath, exists = remoteStore.GetSnippet(resource)
			}
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
		destFileName := baseName + ext
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
		// Try finding matching stacks by base name
		matches := findMatchingRemoteStacks(remoteStore, baseName)
		if len(matches) == 0 {
			return fmt.Errorf("resource '%s' not found in remote registry", resource)
		}

		// Auto-select if only one match
		if len(matches) == 1 {
			resource = matches[0]
			remotePath, exists = remoteStore.GetStack(resource)
		} else {
			// Multiple versions - prompt user
			fmt.Printf("Multiple versions found for '%s':\n", baseName)
			for i, name := range matches {
				fmt.Printf("  %d. %s\n", i+1, name)
			}

			choice, err := utils.Prompt(fmt.Sprintf("Enter version number (1-%d): ", len(matches)))
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			var selectedIdx int
			fmt.Sscanf(choice, "%d", &selectedIdx)
			if selectedIdx < 1 || selectedIdx > len(matches) {
				return fmt.Errorf("invalid choice")
			}

			resource = matches[selectedIdx-1]
			remotePath, exists = remoteStore.GetStack(resource)
		}
	}

	// Check if it's a remote path
	if !store.IsRemotePath(remotePath) {
		return fmt.Errorf("stack '%s' does not have a valid remote location", resource)
	}

	// Download stack to local store first
	localStackPath := filepath.Join(cfg.Paths.Stacks, baseName)

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

	// Copy to destination
	ignorePatterns := []string{"node_modules", ".git", ".DS_Store", "Thumbs.db"}
	if err := utils.CopyDir(localStackPath, destPath, ignorePatterns); err != nil {
		return fmt.Errorf("failed to copy stack: %w", err)
	}

	fmt.Printf(utils.MsgStackAdded, resource, destPath)
	logger.Info(fmt.Sprintf("Remote stack added: %s -> %s", resource, destPath))
	return nil
}

// addDirectRemoteResource adds a resource directly from GitHub without registry lookup
// Supports formats: owner/repo, owner/repo:path, owner/repo@branch:path
func addDirectRemoteResource(remotePath string) error {
	destPath := addTo
	if destPath == "" {
		destPath = "."
	}

	// Parse remote path
	owner, repo, subPath := store.ParseRemotePath(remotePath)
	if owner == "" || repo == "" {
		return fmt.Errorf("invalid remote path format: %s (expected: owner/repo or owner/repo:path)", remotePath)
	}

	fmt.Printf("📥 Fetching directly from %s/%s...\n", owner, repo)

	// Determine if it's a stack (no extension in subPath) or snippet (has extension)
	isSnippet := filepath.Ext(subPath) != ""

	if isSnippet {
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

	// Copy to destination
	ignorePatterns := []string{"node_modules", ".git", ".DS_Store", "Thumbs.db"}
	if err := utils.CopyDir(localStackPath, destPath, ignorePatterns); err != nil {
		return fmt.Errorf("failed to copy stack: %w", err)
	}

	fmt.Printf(utils.MsgStackAdded, resourceName, destPath)
	logger.Info(fmt.Sprintf("Direct remote stack added: %s -> %s", remotePath, destPath))
	return nil
}

// findMatchingRemoteStacks finds all stacks matching the base name in remote store
func findMatchingRemoteStacks(remoteStore *store.Store, baseName string) []string {
	allStacks := remoteStore.ListStacks()
	var matches []string

	for _, stackName := range allStacks {
		stackBase, _, _ := store.ParseResourceName(stackName)
		if stackBase == baseName {
			matches = append(matches, stackName)
		}
	}

	return matches
}

// findMatchingRemoteSnippets finds all snippets matching base name and extension in remote store
func findMatchingRemoteSnippets(remoteStore *store.Store, baseName, ext string) []string {
	allSnippets := remoteStore.ListSnippets()
	var matches []string

	for _, snippetName := range allSnippets {
		snippetBase, _, snippetExt := store.ParseResourceName(snippetName)
		if snippetBase == baseName && snippetExt == ext {
			matches = append(matches, snippetName)
		}
	}

	return matches
}

var (
	addRemote   bool
	addTo       string
	addForce    bool
	addRegistry string
)

func init() {
	addCmd.Flags().BoolVarP(&addRemote, "remote", "r", false, "Fetch from remote registry")
	addCmd.Flags().StringVarP(&addTo, "to", "t", ".", "Destination path")
	addCmd.Flags().BoolVarP(&addForce, FlagForce, FlagForceShort, false, DescForce)
	addCmd.Flags().StringVar(&addRegistry, "registry", "", "Custom registry URL (overrides config)")
}
