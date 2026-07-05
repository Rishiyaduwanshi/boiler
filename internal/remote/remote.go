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
	p := Detect(r.registryURL)
	owner, repo := parseOwnerRepo(r.registryURL)
	ref := resolveProviderRef(p, owner, repo, defaultRemoteRef)
	metaURL := p.RawFileURL(owner, repo, ref, "store/boiler.meta.json")

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
