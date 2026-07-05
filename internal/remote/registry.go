package remote

import (
	"fmt"
	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/store"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

// LoadRegistry resolves the registry URL, expands :VAR references,
// and fetches the remote boiler.meta.json.
func LoadRegistry(registryURL string, cfg *config.Config) (*RemoteStore, *store.Store, error) {
	if registryURL == "" {
		registryURL = cfg.Registry
	}

	resolved, err := utils.ResolveInputToken(registryURL, "registry", cfg.Vars)
	if err != nil {
		return nil, nil, err
	}

	handler, err := NewRemoteStore(resolved)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize remote store: %w", err)
	}

	fmt.Println("🔄 Fetching remote registry...")
	remoteStore, err := handler.LoadFromURL()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load remote registry: %w", err)
	}

	return handler, remoteStore, nil
}
