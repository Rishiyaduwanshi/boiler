package cli

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

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
- Use optional [destination] to override the default path.
- bl add logger .
- bl add logger ./src/utils
- bl add logger /absolute/path

Stack placement:
- By default, stacks are copied inside a stack-named folder.
- bl add express@1 -> ./boiler/express
- Use --spread to copy stack contents directly into destination.
- bl add express@1 --spread -> contents in ./boiler

The command copies resources from your store. For snippets with a single version,
you can use just the name (e.g., 'errorHandler' will auto-select version 1).
For multiple versions, you'll be prompted to choose.

Template Variables:
- Snippets can contain template variables using the format: bl__VAR_NAME.
- When adding a snippet with variables, you'll be prompted to provide values.
- If matching config vars exist, they are used as defaults.
- Default values are shown in brackets (from __var declarations).
- Press Enter to use default or type a custom value.
- Variables are replaced and metadata comments are removed in the final file.

Command Variables:
- Use :name to resolve values from config vars (set via 'bl var').
- Example: bl add express@1 -r --registry :team_reg

Stacks are also versioned and can be added by name or with explicit version.

Remote Resources:
- Use -r flag to fetch from remote source and save to local store.
- Provider is auto-detected from the URL (GitHub, GitLab, Bitbucket, generic).
- Resource is cached locally; subsequent uses do not need -r.
- Use --no-store for one-shot fetch without saving to local store.
- Use --stack/-k or --snippet/-n to override stack/snippet auto-detection.
- For ambiguous remote inputs, stack detection is preferred by default.

Supported remote formats:
- Registry: bl add express@1 -r (registry set via: bl conf --set-registry <url>)
- GitHub short: bl add owner/repo -r
- GitHub full URL: bl add https://github.com/owner/repo -r
- GitLab: bl add https://gitlab.com/owner/repo -r
- Bitbucket: bl add https://bitbucket.org/owner/repo -r
- File from repo: bl add owner/repo:path/to/file.js -r
- Direct file URL: bl add https://site.com/file.js -r
- Direct archive: bl add https://site.com/stack.zip -r
- Custom domain file: bl add site.com:path/file.js -r
- One-time registry: bl add express@1 -r --registry https://github.com/other/boiler`,
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

	# Remote: force snippet mode when auto-detection is unclear
	bl add owner/repo:path/to/template -r --snippet

	# Remote: force stack mode
	bl add https://example.com/custom-source -r --stack

  # Remote: one-time registry override
  bl add express@1 -r --registry https://github.com/myorg/boiler

	# Remote: registry from config variable
	bl add express@1 -r --registry :team_reg

	# One-shot fetch without saving to store
	bl add alice/my-stack -r --no-store`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		resource, err := utils.ResolveInputToken(args[0], "resource", cfg.Vars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		positionalDest := ""
		if len(args) == 2 {
			positionalDest, err = utils.ResolveInputToken(args[1], "destination", cfg.Vars)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		destPath := resolveAddDestination(positionalDest)
		logger.Info(fmt.Sprintf("Adding resource: %s -> %s", resource, destPath))

		if addNoStore && !addRemote {
			fmt.Fprintf(os.Stderr, "Error: --no-store requires --remote\n")
			os.Exit(1)
		}

		if err := addResource(resource, destPath, addRemote, addNoStore); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

const addDefaultDestination = "boiler"

const (
	addForceFlag        = "force"
	addForceFlagShort   = "f"
	addForceDesc        = "Force operation without confirmation"
	addStackFlag        = "stack"
	addStackFlagShort   = "k"
	addStackDesc        = "Treat resource as stack (overrides auto-detection)"
	addSnippetFlag      = "snippet"
	addSnippetFlagShort = "n"
	addSnippetDesc      = "Treat resource as snippet (overrides auto-detection)"
)

type addResourceType int

const (
	addResourceTypeAuto addResourceType = iota
	addResourceTypeStack
	addResourceTypeSnippet
)

func resolveAddDestination(positionalDest string) string {
	if positionalDest == "" {
		return addDefaultDestination
	}
	return filepath.Clean(positionalDest)
}

func resolveAddResourceType() (addResourceType, error) {
	if addAsStack && addAsSnippet {
		return addResourceTypeAuto, fmt.Errorf("--%s and --%s cannot be used together", addStackFlag, addSnippetFlag)
	}

	if addAsStack {
		return addResourceTypeStack, nil
	}

	if addAsSnippet {
		return addResourceTypeSnippet, nil
	}

	return addResourceTypeAuto, nil
}

func resolveSnippetResource(st *store.Store, resource string) (string, error) {
	baseName, version, ext := store.ParseResourceName(resource)

	if version != "" && ext != "" {
		return baseName + "@" + version + ext, nil
	}

	if _, ok := st.GetSnippet(resource); ok {
		return resource, nil
	}

	matches := utils.FindMatchingResources(st.ListSnippets(), baseName, ext)
	if len(matches) == 0 {
		return "", fmt.Errorf(utils.ErrResourceNotFound, "snippet", resource)
	}

	lookupName := baseName
	if ext != "" {
		lookupName = baseName + ext
	}

	selected, err := utils.PickFromList(lookupName, matches)
	if err != nil {
		return "", err
	}

	return selected, nil
}

func resolveStackResource(st *store.Store, resource string) (string, error) {
	baseName, _, _ := store.ParseResourceName(resource)

	if _, ok := st.GetStack(resource); ok {
		return resource, nil
	}

	matches := utils.FindMatchingResources(st.ListStacks(), baseName, "")
	if len(matches) == 0 {
		return "", fmt.Errorf(utils.ErrResourceNotFound, "stack", resource)
	}

	selected, err := utils.PickFromList(baseName, matches)
	if err != nil {
		return "", err
	}

	return selected, nil
}

func addResource(resource, destPath string, remoteEnabled bool, noStore bool) error {
	resourceType, err := resolveAddResourceType()
	if err != nil {
		return err
	}

	if remoteEnabled {
		return addResourceFromRemote(resource, destPath, resourceType, noStore)
	}

	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}

	if resourceType == addResourceTypeSnippet {
		if addSpread {
			return fmt.Errorf("--spread is only supported for stacks")
		}

		selected, err := resolveSnippetResource(st, resource)
		if err != nil {
			return err
		}
		return addSnippet(st, selected, destPath)
	}

	if resourceType == addResourceTypeStack {
		selected, err := resolveStackResource(st, resource)
		if err != nil {
			return err
		}
		return addStack(st, selected, destPath)
	}

	baseName, _, ext := store.ParseResourceName(resource)

	if ext != "" {
		if addSpread {
			return fmt.Errorf("--spread is only supported for stacks")
		}

		selected, err := resolveSnippetResource(st, resource)
		if err != nil {
			return err
		}
		return addSnippet(st, selected, destPath)
	}

	selectedStack, stackErr := resolveStackResource(st, resource)
	if stackErr == nil {
		return addStack(st, selectedStack, destPath)
	}

	matches := utils.FindMatchingResources(st.ListSnippets(), baseName, "")
	if len(matches) == 0 {
		return fmt.Errorf(utils.ErrResourceNotFound, "stack or snippet", resource)
	}

	selectedSnippet, err := utils.PickFromList(baseName, matches)
	if err != nil {
		return err
	}

	return addSnippet(st, selectedSnippet, destPath)
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
func addResourceFromRemote(resource, destPath string, resourceType addResourceType, noStore bool) error {
	// Check if resource is direct GitHub path (owner/repo format)
	if store.IsRemotePath(resource) {
		return addDirectRemoteResource(resource, destPath, resourceType, noStore)
	}

	// Use custom registry if provided, otherwise use config
	registryURL := cfg.Registry
	if addRegistry != "" {
		registryURL = addRegistry
	}

	resolvedRegistryURL, err := utils.ResolveInputToken(registryURL, "registry", cfg.Vars)
	if err != nil {
		return err
	}
	registryURL = resolvedRegistryURL

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

	baseName, _, ext := store.ParseResourceName(resource)
	fetchAsSnippet := ext != ""

	if resourceType == addResourceTypeSnippet {
		fetchAsSnippet = true
	}

	if resourceType == addResourceTypeStack {
		fetchAsSnippet = false
	}

	if fetchAsSnippet {
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

		// Parse remote path to get snippet file location inside repo/domain.
		_, _, remoteFile := store.ParseRemotePath(remotePath)
		if remoteFile == "" || remoteFile == "." {
			return fmt.Errorf("invalid remote snippet path: %s", remotePath)
		}

		if noStore {
			baseName, _, snippetExt := store.ParseResourceName(resource)
			targetFileName := baseName + snippetExt
			if targetFileName == "" {
				targetFileName = filepath.Base(remoteFile)
			}
			destFile := filepath.Join(destPath, targetFileName)
			if utils.FileExists(destFile) && !addForce {
				return fmt.Errorf(utils.ErrFileAlreadyExists, destFile)
			}

			if err := remote.FetchSnippet(remotePath, destFile); err != nil {
				return err
			}

			fmt.Printf(utils.MsgSnippetAdded, resource, destFile)
			logger.Info(fmt.Sprintf("Remote snippet added (no-store): %s -> %s", resource, destFile))
			return nil
		}

		localRelativePath := filepath.Clean(filepath.FromSlash(remoteFile))
		if localRelativePath == "." || filepath.IsAbs(localRelativePath) || strings.HasPrefix(localRelativePath, "..") {
			return fmt.Errorf("invalid remote snippet path: %s", remotePath)
		}

		localDestPath := filepath.Join(cfg.Paths.Snippets, localRelativePath)
		if err := utils.EnsureDir(filepath.Dir(localDestPath)); err != nil {
			return fmt.Errorf("failed to create local store directory: %w", err)
		}

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

		return addSnippet(st, resource, destPath)
	}

	if resourceType == addResourceTypeSnippet {
		return fmt.Errorf("snippet '%s' not found in remote registry", resource)
	}

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

	if noStore {
		tempStackPath, err := os.MkdirTemp("", "boiler-remote-stack-*")
		if err != nil {
			return fmt.Errorf("failed to create temp stack directory: %w", err)
		}
		defer os.RemoveAll(tempStackPath)

		if err := remote.FetchStack(remotePath, tempStackPath); err != nil {
			return err
		}

		finalDestPath, err := copyStackToDestination(tempStackPath, resource, destPath)
		if err != nil {
			return err
		}

		fmt.Printf(utils.MsgStackAdded, resource, finalDestPath)
		logger.Info(fmt.Sprintf("Remote stack added (no-store): %s -> %s", resource, finalDestPath))
		return nil
	}

	// Download stack to local store first
	localStackPath := filepath.Join(cfg.Paths.Stacks, stackDirectoryName(resource))
	if err := os.RemoveAll(localStackPath); err != nil {
		return fmt.Errorf("failed to reset local stack cache: %w", err)
	}

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

// addDirectRemoteResource adds a resource directly from remote without registry lookup.
func shouldTreatDirectRemotePathAsSnippet(subPath string, resourceType addResourceType) bool {
	switch resourceType {
	case addResourceTypeSnippet:
		return true
	case addResourceTypeStack:
		return false
	default:
		return filepath.Ext(subPath) != ""
	}
}

func shouldTreatDirectURLAsSnippet(resource string, resourceType addResourceType) bool {
	switch resourceType {
	case addResourceTypeSnippet:
		return true
	case addResourceTypeStack:
		return false
	default:
		return isDirectRemoteFileURL(resource)
	}
}

func addDirectRemoteResource(remotePath, destPath string, resourceType addResourceType, noStore bool) error {
	if isHTTPRemotePath(remotePath) {
		return addDirectRemoteURLResource(remotePath, destPath, resourceType, noStore)
	}

	// Parse remote path
	owner, repo, subPath := store.ParseRemotePath(remotePath)
	if owner == "" || repo == "" {
		return fmt.Errorf("invalid remote path format: %s (expected: owner/repo, owner/repo:path, provider:owner/repo/path, or full URL)", remotePath)
	}

	fmt.Printf("📥 Fetching directly from %s/%s...\n", owner, repo)

	isSnippet := shouldTreatDirectRemotePathAsSnippet(subPath, resourceType)

	if isSnippet {
		if addSpread {
			return fmt.Errorf("--spread is only supported for stacks")
		}

		if subPath == "" || subPath == "." || strings.HasSuffix(subPath, "/") {
			return fmt.Errorf("snippet path must point to a file: %s", remotePath)
		}

		if noStore {
			resourceName := filepath.Base(subPath)
			destFile := filepath.Join(destPath, resourceName)
			if utils.FileExists(destFile) && !addForce {
				return fmt.Errorf(utils.ErrFileAlreadyExists, destFile)
			}

			if err := remote.FetchSnippet(remotePath, destFile); err != nil {
				return err
			}

			fmt.Printf(utils.MsgSnippetAdded, resourceName, destFile)
			logger.Info(fmt.Sprintf("Direct remote snippet added (no-store): %s -> %s", remotePath, destFile))
			return nil
		}

		// Download snippet - use extension-based directory to match bl store behavior
		ext := strings.TrimPrefix(filepath.Ext(filepath.Base(subPath)), ".")
		langDir := filepath.Join(cfg.Paths.Snippets, ext)
		if err := os.MkdirAll(langDir, 0755); err != nil {
			return fmt.Errorf("failed to create local store directory: %w", err)
		}

		localDestPath := filepath.Join(langDir, filepath.Base(subPath))

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

		return addSnippet(st, resourceName, destPath)
	}

	if noStore {
		tempStackPath, err := os.MkdirTemp("", "boiler-direct-stack-*")
		if err != nil {
			return fmt.Errorf("failed to create temp stack directory: %w", err)
		}
		defer os.RemoveAll(tempStackPath)

		if err := remote.FetchStack(remotePath, tempStackPath); err != nil {
			return err
		}

		resourceName := repo
		finalDestPath, err := copyStackToDestination(tempStackPath, resourceName, destPath)
		if err != nil {
			return err
		}

		fmt.Printf(utils.MsgStackAdded, resourceName, finalDestPath)
		logger.Info(fmt.Sprintf("Direct remote stack added (no-store): %s -> %s", remotePath, finalDestPath))
		return nil
	}

	// It's a stack
	localStackPath := filepath.Join(cfg.Paths.Stacks, repo)
	if err := os.RemoveAll(localStackPath); err != nil {
		return fmt.Errorf("failed to reset local stack cache: %w", err)
	}

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

func isHTTPRemotePath(resource string) bool {
	return utils.IsURL(resource)
}

func isDirectRemoteFileURL(resource string) bool {
	lower := strings.ToLower(resource)
	trimmed := strings.SplitN(resource, "?", 2)[0]
	trimmed = strings.SplitN(trimmed, "#", 2)[0]

	parsed, err := url.Parse(trimmed)
	if err == nil {
		host := strings.ToLower(parsed.Hostname())
		segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")

		switch host {
		case "github.com", "www.github.com":
			if len(segments) >= 4 && segments[2] == "blob" {
				// Blob URLs without extension are most likely directory-like inputs.
				if filepath.Ext(segments[len(segments)-1]) == "" {
					return false
				}
				return true
			}
		case "gitlab.com", "www.gitlab.com":
			if len(segments) >= 5 && segments[2] == "-" && segments[3] == "blob" {
				if filepath.Ext(segments[len(segments)-1]) == "" {
					return false
				}
				return true
			}
		}
	}

	if strings.Contains(lower, "raw.githubusercontent.com/") ||
		strings.Contains(lower, "/-/raw/") {
		return true
	}

	switch strings.ToLower(filepath.Ext(trimmed)) {
	case "", ".zip", ".tar", ".gz", ".tgz":
		return false
	default:
		return true
	}
}

func fileNameFromRemoteURL(remotePath string) string {
	parsed, err := url.Parse(remotePath)
	if err == nil {
		name := path.Base(parsed.Path)
		if name != "" && name != "." && name != "/" {
			return name
		}
	}

	name := filepath.Base(remotePath)
	if name == "" || name == "." || name == "/" {
		return ""
	}
	return name
}

func trimArchiveSuffix(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"):
		return name[:len(name)-len(".tar.gz")]
	case strings.HasSuffix(lower, ".tgz"):
		return name[:len(name)-len(".tgz")]
	case strings.HasSuffix(lower, ".zip"):
		return name[:len(name)-len(".zip")]
	default:
		return name
	}
}

func stackNameFromRemoteURL(remotePath string) string {
	parsed, err := url.Parse(remotePath)
	if err == nil {
		host := strings.ToLower(parsed.Hostname())
		segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")

		if (host == "github.com" || host == "gitlab.com" || host == "bitbucket.org") && len(segments) >= 2 {
			repo := strings.TrimSuffix(segments[1], ".git")
			if repo != "" {
				return repo
			}
		}

		if len(segments) > 0 {
			name := trimArchiveSuffix(segments[len(segments)-1])
			if name != "" && name != "." && name != "/" {
				return name
			}
		}
	}

	return "remote-stack"
}

func addDirectRemoteURLResource(remotePath, destPath string, resourceType addResourceType, noStore bool) error {
	if shouldTreatDirectURLAsSnippet(remotePath, resourceType) {
		if addSpread {
			return fmt.Errorf("--spread is only supported for stacks")
		}

		resourceName := fileNameFromRemoteURL(remotePath)
		if resourceName == "" {
			return fmt.Errorf("could not determine snippet file name from URL: %s", remotePath)
		}

		if noStore {
			destFile := filepath.Join(destPath, resourceName)
			if utils.FileExists(destFile) && !addForce {
				return fmt.Errorf(utils.ErrFileAlreadyExists, destFile)
			}
			if err := remote.FetchSnippet(remotePath, destFile); err != nil {
				return err
			}
			fmt.Printf(utils.MsgSnippetAdded, resourceName, destFile)
			logger.Info(fmt.Sprintf("Direct URL snippet added (no-store): %s -> %s", remotePath, destFile))
			return nil
		}

		// Use extension-based directory to match bl store behavior
		ext := strings.TrimPrefix(filepath.Ext(resourceName), ".")
		langDir := filepath.Join(cfg.Paths.Snippets, ext)
		if err := utils.EnsureDir(langDir); err != nil {
			return fmt.Errorf("failed to create local store directory: %w", err)
		}
		localDestPath := filepath.Join(langDir, resourceName)

		if err := remote.FetchSnippet(remotePath, localDestPath); err != nil {
			return err
		}

		st, err := utils.LoadStore(cfg.Paths.Store)
		if err != nil {
			return err
		}
		if err := st.AddSnippet(resourceName, localDestPath); err != nil {
			return err
		}

		return addSnippet(st, resourceName, destPath)
	}

	resourceName := stackNameFromRemoteURL(remotePath)
	if noStore {
		tempStackPath, err := os.MkdirTemp("", "boiler-url-stack-*")
		if err != nil {
			return fmt.Errorf("failed to create temp stack directory: %w", err)
		}
		defer os.RemoveAll(tempStackPath)

		if err := remote.FetchStack(remotePath, tempStackPath); err != nil {
			return err
		}

		finalDestPath, err := copyStackToDestination(tempStackPath, resourceName, destPath)
		if err != nil {
			return err
		}

		fmt.Printf(utils.MsgStackAdded, resourceName, finalDestPath)
		logger.Info(fmt.Sprintf("Direct URL stack added (no-store): %s -> %s", remotePath, finalDestPath))
		return nil
	}

	localStackPath := filepath.Join(cfg.Paths.Stacks, stackDirectoryName(resourceName))
	if err := os.RemoveAll(localStackPath); err != nil {
		return fmt.Errorf("failed to reset local stack cache: %w", err)
	}

	if err := remote.FetchStack(remotePath, localStackPath); err != nil {
		return err
	}

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
	addRemote    bool
	addSpread    bool
	addForce     bool
	addRegistry  string
	addAsStack   bool
	addAsSnippet bool
	addNoStore   bool
)

func init() {
	addCmd.Flags().BoolVarP(&addRemote, "remote", "r", false, "Fetch from remote registry")
	addCmd.Flags().BoolVar(&addNoStore, "no-store", false, "Fetch remote resource without saving to local store")
	addCmd.Flags().BoolVar(&addSpread, "spread", false, "Spread stack contents directly into destination")
	addCmd.Flags().BoolVarP(&addForce, addForceFlag, addForceFlagShort, false, addForceDesc)
	addCmd.Flags().BoolVarP(&addAsStack, addStackFlag, addStackFlagShort, false, addStackDesc)
	addCmd.Flags().BoolVarP(&addAsSnippet, addSnippetFlag, addSnippetFlagShort, false, addSnippetDesc)
	addCmd.Flags().StringVar(&addRegistry, "registry", "", "Custom registry URL (overrides config)")
}
