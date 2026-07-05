package add

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/remote"
	"github.com/rishiyaduwanshi/boiler/internal/store"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

// ResourceFromRemote fetches and adds a resource from remote registry
func ResourceFromRemote(resource, destPath string, resourceType ResourceType, noStore bool, opts Options, cfg *config.Config, logger *utils.Logger) error {
	// Check if resource is direct GitHub path (owner/repo format)
	if store.IsRemotePath(resource) {
		return directRemoteResource(resource, destPath, resourceType, noStore, opts, cfg, logger)
	}

	_, remoteStore, err := remote.LoadRegistry(opts.Registry, cfg)
	if err != nil {
		return err
	}

	baseName, _, ext := store.ParseResourceName(resource)
	fetchAsSnippet := ext != ""

	if resourceType == ResourceTypeSnippet {
		fetchAsSnippet = true
	}

	if resourceType == ResourceTypeStack {
		fetchAsSnippet = false
	}

	if fetchAsSnippet {
		if opts.Spread {
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
			if utils.FileExists(destFile) && !opts.Force {
				return fmt.Errorf(utils.ErrFileAlreadyExists, destFile)
			}

			if err := remote.FetchSnippet(remotePath, destFile); err != nil {
				return err
			}

			fmt.Printf(utils.MsgSnippetAdded, resource, destFile)
			if logger != nil {
				logger.Info(fmt.Sprintf("Remote snippet added (no-store): %s -> %s", resource, destFile))
			}
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

		return AddSnippet(st, resource, destPath, opts, cfg, logger)
	}

	if resourceType == ResourceTypeSnippet {
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

		finalDestPath, err := copyStackToDestination(tempStackPath, resource, destPath, opts)
		if err != nil {
			return err
		}

		fmt.Printf(utils.MsgStackAdded, resource, finalDestPath)
		if logger != nil {
			logger.Info(fmt.Sprintf("Remote stack added (no-store): %s -> %s", resource, finalDestPath))
		}
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

	finalDestPath, err := copyStackToDestination(localStackPath, resource, destPath, opts)
	if err != nil {
		return err
	}

	fmt.Printf(utils.MsgStackAdded, resource, finalDestPath)
	if logger != nil {
		logger.Info(fmt.Sprintf("Remote stack added: %s -> %s", resource, finalDestPath))
	}
	return nil
}

// shouldTreatDirectRemotePathAsSnippet reports whether a short remote path
// (owner/repo:subpath) should be fetched as a snippet.
func shouldTreatDirectRemotePathAsSnippet(subPath string, resourceType ResourceType) bool {
	switch resourceType {
	case ResourceTypeSnippet:
		return true
	case ResourceTypeStack:
		return false
	default:
		return filepath.Ext(subPath) != ""
	}
}

// shouldTreatDirectURLAsSnippet reports whether a full URL should be fetched
// as a snippet based on the URL structure and the explicit resource type flag.
func shouldTreatDirectURLAsSnippet(resource string, resourceType ResourceType) bool {
	switch resourceType {
	case ResourceTypeSnippet:
		return true
	case ResourceTypeStack:
		return false
	default:
		return utils.IsDirectRemoteFileURL(resource)
	}
}

func directRemoteResource(remotePath, destPath string, resourceType ResourceType, noStore bool, opts Options, cfg *config.Config, logger *utils.Logger) error {
	if utils.IsURL(remotePath) {
		return directRemoteURLResource(remotePath, destPath, resourceType, noStore, opts, cfg, logger)
	}

	// Parse remote path
	owner, repo, subPath := store.ParseRemotePath(remotePath)
	if owner == "" || repo == "" {
		return fmt.Errorf("invalid remote path format: %s (expected: owner/repo, owner/repo:path, provider:owner/repo/path, or full URL)", remotePath)
	}

	fmt.Printf("📥 Fetching directly from %s/%s...\n", owner, repo)

	isSnippet := shouldTreatDirectRemotePathAsSnippet(subPath, resourceType)

	if isSnippet {
		if opts.Spread {
			return fmt.Errorf("--spread is only supported for stacks")
		}

		if subPath == "" || subPath == "." || strings.HasSuffix(subPath, "/") {
			return fmt.Errorf("snippet path must point to a file: %s", remotePath)
		}

		if noStore {
			resourceName := filepath.Base(subPath)
			destFile := filepath.Join(destPath, resourceName)
			if utils.FileExists(destFile) && !opts.Force {
				return fmt.Errorf(utils.ErrFileAlreadyExists, destFile)
			}

			if err := remote.FetchSnippet(remotePath, destFile); err != nil {
				return err
			}

			fmt.Printf(utils.MsgSnippetAdded, resourceName, destFile)
			if logger != nil {
				logger.Info(fmt.Sprintf("Direct remote snippet added (no-store): %s -> %s", remotePath, destFile))
			}
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

		return AddSnippet(st, resourceName, destPath, opts, cfg, logger)
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
		finalDestPath, err := copyStackToDestination(tempStackPath, resourceName, destPath, opts)
		if err != nil {
			return err
		}

		fmt.Printf(utils.MsgStackAdded, resourceName, finalDestPath)
		if logger != nil {
			logger.Info(fmt.Sprintf("Direct remote stack added (no-store): %s -> %s", remotePath, finalDestPath))
		}
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

	finalDestPath, err := copyStackToDestination(localStackPath, resourceName, destPath, opts)
	if err != nil {
		return err
	}

	fmt.Printf(utils.MsgStackAdded, resourceName, finalDestPath)
	if logger != nil {
		logger.Info(fmt.Sprintf("Direct remote stack added: %s -> %s", remotePath, finalDestPath))
	}
	return nil
}

func directRemoteURLResource(remotePath, destPath string, resourceType ResourceType, noStore bool, opts Options, cfg *config.Config, logger *utils.Logger) error {
	if shouldTreatDirectURLAsSnippet(remotePath, resourceType) {
		if opts.Spread {
			return fmt.Errorf("--spread is only supported for stacks")
		}

		resourceName := utils.FileNameFromRemoteURL(remotePath)
		if resourceName == "" {
			return fmt.Errorf("could not determine snippet file name from URL: %s", remotePath)
		}

		if noStore {
			destFile := filepath.Join(destPath, resourceName)
			if utils.FileExists(destFile) && !opts.Force {
				return fmt.Errorf(utils.ErrFileAlreadyExists, destFile)
			}
			if err := remote.FetchSnippet(remotePath, destFile); err != nil {
				return err
			}
			fmt.Printf(utils.MsgSnippetAdded, resourceName, destFile)
			if logger != nil {
				logger.Info(fmt.Sprintf("Direct URL snippet added (no-store): %s -> %s", remotePath, destFile))
			}
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

		return AddSnippet(st, resourceName, destPath, opts, cfg, logger)
	}

	resourceName := utils.StackNameFromRemoteURL(remotePath)
	if noStore {
		tempStackPath, err := os.MkdirTemp("", "boiler-url-stack-*")
		if err != nil {
			return fmt.Errorf("failed to create temp stack directory: %w", err)
		}
		defer os.RemoveAll(tempStackPath)

		if err := remote.FetchStack(remotePath, tempStackPath); err != nil {
			return err
		}

		finalDestPath, err := copyStackToDestination(tempStackPath, resourceName, destPath, opts)
		if err != nil {
			return err
		}

		fmt.Printf(utils.MsgStackAdded, resourceName, finalDestPath)
		if logger != nil {
			logger.Info(fmt.Sprintf("Direct URL stack added (no-store): %s -> %s", remotePath, finalDestPath))
		}
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

	finalDestPath, err := copyStackToDestination(localStackPath, resourceName, destPath, opts)
	if err != nil {
		return err
	}

	fmt.Printf(utils.MsgStackAdded, resourceName, finalDestPath)
	if logger != nil {
		logger.Info(fmt.Sprintf("Direct remote stack added: %s -> %s", remotePath, finalDestPath))
	}
	return nil
}
