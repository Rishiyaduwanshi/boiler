package remote

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/store"
)

// RemoteStore wraps store.Store to handle remote operations
type RemoteStore struct {
	registryURL string
	store       *store.Store
}

// NewRemoteStore creates a new remote store instance
func NewRemoteStore(registryURL string) (*RemoteStore, error) {
	return &RemoteStore{
		registryURL: registryURL,
	}, nil
}

// LoadFromURL fetches and loads the remote boiler.meta.json
func (r *RemoteStore) LoadFromURL() (*store.Store, error) {
	// Construct URL to boiler.meta.json in remote registry
	// Example: https://raw.githubusercontent.com/rishiyaduwanshi/boiler/main/store/boiler.meta.json
	metaURL := buildRawURL(r.registryURL, "store/boiler.meta.json")

	// Download metadata
	data, err := downloadFile(metaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote metadata: %w", err)
	}

	// Create a temporary store to parse the metadata
	st := &store.Store{}
	meta := &store.Meta{
		Stacks:   make(map[string]string),
		Snippets: make(map[string]string),
	}
	
	// Parse json.Unmarshal
	if err := json.Unmarshal(data, meta); err != nil {
		return nil, fmt.Errorf("failed to parse remote metadata: %w", err)
	}

	// Set the metadata
	st.SetMeta(meta)
	r.store = st

	return st, nil
}

// Search searches for resources matching query
func (r *RemoteStore) Search(st *store.Store, query string, searchSnippets, searchStacks bool) map[string][]string {
	results := make(map[string][]string)
	query = strings.ToLower(query)
	
	if searchSnippets {
		snippets := []string{}
		for _, name := range st.ListSnippets() {
			if strings.Contains(strings.ToLower(name), query) {
				snippets = append(snippets, name)
			}
		}
		if len(snippets) > 0 {
			results["snippets"] = snippets
		}
	}

	if searchStacks {
		stacks := []string{}
		for _, name := range st.ListStacks() {
			if strings.Contains(strings.ToLower(name), query) {
				stacks = append(stacks, name)
			}
		}
		if len(stacks) > 0 {
			results["stacks"] = stacks
		}
	}

	return results
}

// Helper functions

// buildRawURL constructs GitHub raw content URL from registry URL
// Input: "https://github.com/owner/repo", "store/boiler.meta.json"
// Output: "https://raw.githubusercontent.com/owner/repo/main/store/boiler.meta.json"
func buildRawURL(registryURL, path string) string {
	// Extract owner and repo from registry URL
	// https://github.com/rishiyaduwanshi/boiler -> rishiyaduwanshi/boiler
	registryURL = strings.TrimPrefix(registryURL, "https://github.com/")
	registryURL = strings.TrimPrefix(registryURL, "http://github.com/")
	registryURL = strings.TrimSuffix(registryURL, "/")
	
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s", registryURL, path)
}
