package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/constants"
)

type Meta struct {
	Stacks   map[string]string `json:"stacks"`
	Snippets map[string]string `json:"snippets"`
}

type Store struct {
	metaPath string
	meta     *Meta
}

func NewStore(storePath string) *Store {
	return &Store{
		metaPath: filepath.Join(storePath, "boiler.meta.json"),
		meta: &Meta{
			Stacks:   make(map[string]string),
			Snippets: make(map[string]string),
		},
	}
}

func (s *Store) Load() error {
	data, err := os.ReadFile(s.metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.Save()
		}
		return fmt.Errorf("failed to read meta file: %w", err)
	}

	if err := json.Unmarshal(data, s.meta); err != nil {
		return fmt.Errorf("failed to parse meta file: %w", err)
	}

	if s.meta.Stacks == nil {
		s.meta.Stacks = make(map[string]string)
	}
	if s.meta.Snippets == nil {
		s.meta.Snippets = make(map[string]string)
	}

	return nil
}

func (s *Store) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.metaPath), 0755); err != nil {
		return fmt.Errorf("failed to create meta directory: %w", err)
	}

	data, err := json.MarshalIndent(s.meta, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	if err := os.WriteFile(s.metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write meta file: %w", err)
	}

	return nil
}

func (s *Store) SetMeta(meta *Meta) {
	s.meta = meta
}

func (s *Store) GetMeta() *Meta {
	return s.meta
}

func (s *Store) AddSnippet(name, path string) error {
	if rel, err := filepath.Rel(filepath.Dir(s.metaPath), path); err == nil {
		path = rel
	}
	s.meta.Snippets[name] = filepath.ToSlash(path)
	return s.Save()
}

func (s *Store) AddStack(name, path string) error {
	if rel, err := filepath.Rel(filepath.Dir(s.metaPath), path); err == nil {
		path = rel
	}
	s.meta.Stacks[name] = filepath.ToSlash(path)
	return s.Save()
}

func (s *Store) resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(filepath.Dir(s.metaPath), p)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *Store) GetSnippet(name string) (string, bool) {
	path, ok := s.meta.Snippets[name]
	if ok {
		resolved := s.resolvePath(path)
		if fileExists(resolved) {
			return resolved, true
		}
	}

	// O(1) Fallback: Check the extension-based folder directly
	ext := filepath.Ext(name)
	if ext != "" {
		langDir := strings.TrimPrefix(ext, ".")
		snippetsDir := filepath.Join(filepath.Dir(s.metaPath), constants.SnippetsDirName)
		fallbackPath := filepath.Join(snippetsDir, langDir, name)
		if fileExists(fallbackPath) {
			return fallbackPath, true
		}
	}

	return "", false
}

func (s *Store) GetStack(name string) (string, bool) {
	path, ok := s.meta.Stacks[name]
	if ok {
		resolved := s.resolvePath(path)
		if fileExists(resolved) {
			return resolved, true
		}
	}

	// O(1) Fallback: Check the stacks folder directly
	stacksDir := filepath.Join(filepath.Dir(s.metaPath), constants.StacksDirName)
	fallbackPath := filepath.Join(stacksDir, name)
	if fileExists(fallbackPath) {
		return fallbackPath, true
	}

	return "", false
}

func (s *Store) RemoveSnippet(name string) error {
	delete(s.meta.Snippets, name)
	return s.Save()
}

func (s *Store) RemoveStack(name string) error {
	delete(s.meta.Stacks, name)
	return s.Save()
}

func (s *Store) SnippetExists(name string) bool {
	_, ok := s.meta.Snippets[name]
	return ok
}

func (s *Store) StackExists(name string) bool {
	_, ok := s.meta.Stacks[name]
	return ok
}

func (s *Store) ListSnippets() []string {
	snippets := make([]string, 0, len(s.meta.Snippets))
	for name := range s.meta.Snippets {
		snippets = append(snippets, name)
	}
	return snippets
}

func (s *Store) ListStacks() []string {
	stacks := make([]string, 0, len(s.meta.Stacks))
	for name := range s.meta.Stacks {
		stacks = append(stacks, name)
	}
	return stacks
}

// GetAllVersions returns all existing version numbers for a snippet, sorted in ascending order.
// This is used to check if a snippet already exists and provide options to overwrite or create new version.
// Example: For snippets "logger@1.js", "logger@3.js", "logger@5.js", returns []int{1, 3, 5}
// Returns empty slice if no versions exist.
func (s *Store) GetAllVersions(baseName, extension string) []int {
	versions := []int{}

	for snippetName := range s.meta.Snippets {
		name, versionStr, ext := ParseResourceName(snippetName)

		// Match by base name and extension
		if name == baseName && ext == extension {
			if versionStr != "" {
				if version, err := strconv.Atoi(versionStr); err == nil {
					versions = append(versions, version)
				}
			}
		}
	}

	sort.Ints(versions)
	return versions
}

func ParseResourceName(resource string) (name, version, extension string) {
	parts := strings.SplitN(resource, "@", 2)
	nameWithExt := parts[0]

	if len(parts) == 2 {
		versionWithExt := parts[1]
		// Check if version has extension
		ext := filepath.Ext(versionWithExt)
		if ext != "" {
			version = strings.TrimSuffix(versionWithExt, ext)
			extension = ext
		} else {
			version = versionWithExt
		}
	}

	if extension == "" {
		extension = filepath.Ext(nameWithExt)
		name = strings.TrimSuffix(nameWithExt, extension)
	} else {
		name = nameWithExt
	}

	return name, version, extension
}

func IsStack(resource string) bool {
	_, _, ext := ParseResourceName(resource)
	return ext == ""
}

func IsSnippet(resource string) bool {
	return !IsStack(resource)
}

func isProviderPrefix(prefix string) bool {
	switch strings.ToLower(prefix) {
	case constants.ProviderGithubAlias, constants.ProviderGithubHost, constants.ProviderGithubWWW,
		constants.ProviderGitlabAlias, constants.ProviderGitlabHost, constants.ProviderGitlabWWW,
		constants.ProviderBitbucketAlias, constants.ProviderBitbucketHost, constants.ProviderBitbucketWWW:
		return true
	default:
		return false
	}
}

func parseOwnerRepoPath(value string) (owner, repo, subPath string, ok bool) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(value), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 {
		return "", "", "", false
	}

	owner = parts[0]
	repo = parts[1]
	if owner == "" || repo == "" {
		return "", "", "", false
	}

	if len(parts) == 2 {
		return owner, repo, ".", true
	}

	subPath = strings.Join(parts[2:], "/")
	if subPath == "" {
		subPath = "."
	}

	return owner, repo, subPath, true
}

func isURL(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(lower, constants.SchemeHTTP) || strings.HasPrefix(lower, constants.SchemeHTTPS)
}

// IsRemotePath checks if a path is a remote GitHub location or URL
// Formats: "owner/repo:path", "owner/repo", "https://domain.com/path", "domain.com:path"
func IsRemotePath(path string) bool {
	// Check if it's a full URL
	if isURL(path) {
		return true
	}

	// Remote paths contain / for GitHub owner/repo
	// and don't start with drive letters or / (absolute paths)
	if strings.Contains(path, ":") {
		parts := strings.SplitN(path, ":", 2)
		prefix := strings.ToLower(parts[0])

		if isProviderPrefix(prefix) {
			return true
		}

		// Check if it's like "owner/repo:path" or "domain.com:path" not "C:/path"
		if strings.Contains(parts[0], "/") && !strings.HasPrefix(path, "/") {
			return true
		}
		// Check for domain:path format (e.g., "iamabhinav.dev:snippets/file.js")
		if strings.Contains(parts[0], ".") && !strings.HasPrefix(path, "/") {
			return true
		}
	}
	// Check if it's just "owner/repo" format
	if strings.Count(path, "/") == 1 && !strings.HasPrefix(path, "/") && !strings.Contains(path, ":") {
		return true
	}
	return false
}

// ParseRemotePath parses a remote path into owner, repo, and path components
// Format: "owner/repo:path" -> ("owner", "repo", "path")
// Format: "owner/repo" -> ("owner", "repo", ".")
// Format: "https://domain.com/path" -> ("", "https://domain.com/path", "")
// Format: "domain.com:path" -> ("", "domain.com", "path")
func ParseRemotePath(remotePath string) (owner, repo, path string) {
	// Check if it's a full URL
	if isURL(remotePath) {
		// Return the full URL as "repo" (it's actually a direct URL)
		return "", remotePath, ""
	}

	// Check for domain:path format (e.g., "iamabhinav.dev:snippets/file.js")
	if strings.Contains(remotePath, ":") {
		parts := strings.SplitN(remotePath, ":", 2)
		prefix := strings.TrimSpace(parts[0])
		filePath := strings.TrimPrefix(strings.TrimSpace(parts[1]), "/")

		if isProviderPrefix(prefix) {
			if owner, repo, path, ok := parseOwnerRepoPath(filePath); ok {
				return owner, repo, path
			}
			return "", "", ""
		}

		// owner/repo:path format should be resolved before domain:path,
		// because repository names can contain dots (e.g., next.js).
		repoparts := strings.SplitN(prefix, "/", 2)
		if len(repoparts) == 2 {
			return repoparts[0], repoparts[1], filePath
		}

		// If it contains a dot, it's likely a domain
		if strings.Contains(prefix, ".") {
			return "", prefix, filePath
		}
	} else {
		// Check for simple owner/repo format
		parts := strings.SplitN(remotePath, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1], "."
		}
	}
	return "", "", ""
}
