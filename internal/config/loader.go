package config

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

// Load orchestrates finding, loading, and merging Global and Local configurations.
func Load() (*Manager, error) {
	manager := &Manager{
		Global:  &ConfigFile{Exists: false},
		Local:   &ConfigFile{Exists: false},
		Runtime: DefaultConfig(),
	}

	// 1. Load Global Config
	globalPath, err := GlobalConfigPath()
	if err == nil {
		manager.Global.Path = globalPath
		if data, err := os.ReadFile(globalPath); err == nil {
			var globalCfg Config
			if err := json.Unmarshal(data, &globalCfg); err != nil {
				return nil, fmt.Errorf("failed to parse global config at %s: %w", globalPath, err)
			}
			manager.Global.Config = &globalCfg
			manager.Global.Exists = true
		}
	}

	// 2. Load Local Config (Nearest Parent)
	cwd, err := os.Getwd()
	if err == nil {
		localPath, err := FindNearestConfig(cwd)
		if err == nil && localPath != "" {
			manager.Local.Path = localPath
			if data, err := os.ReadFile(localPath); err == nil {
				var localCfg Config
				if err := json.Unmarshal(data, &localCfg); err != nil {
					return nil, fmt.Errorf("failed to parse local config at %s: %w", localPath, err)
				}
				manager.Local.Config = &localCfg
				manager.Local.Exists = true
			}
		} else {
			// Initialize empty local config path for default saving
			manager.Local.Path = DefaultLocalConfigPath(cwd)
		}
	}

	// 3. Merge into Runtime
	mergeIntoRuntime(manager)

	// Ensure runtime defaults are filled
	mergeWithDefaults(manager.Runtime)

	return manager, nil
}

// mergeIntoRuntime applies global config, then overrides with local config
func mergeIntoRuntime(m *Manager) {
	if m.Global.Exists && m.Global.Config != nil {
		m.Runtime.Scope = m.Global.Config.Scope

		if m.Global.Config.Aliases != nil {
			m.Runtime.Aliases = maps.Clone(m.Global.Config.Aliases)
		}
		if m.Global.Config.Vars != nil {
			m.Runtime.Vars = maps.Clone(m.Global.Config.Vars)
		}
		if m.Global.Config.Artifacts != nil {
			m.Runtime.Artifacts = maps.Clone(m.Global.Config.Artifacts)
		}
	}

	if m.Local.Exists && m.Local.Config != nil {
		if m.Local.Config.Scope != "" {
			m.Runtime.Scope = m.Local.Config.Scope
		}

		if m.Local.Config.Aliases != nil {
			if m.Runtime.Aliases == nil {
				m.Runtime.Aliases = make(map[string]string)
			}
			maps.Copy(m.Runtime.Aliases, m.Local.Config.Aliases)
		}

		if m.Local.Config.Vars != nil {
			if m.Runtime.Vars == nil {
				m.Runtime.Vars = make(map[string]string)
			}
			maps.Copy(m.Runtime.Vars, m.Local.Config.Vars)
		}

		if m.Local.Config.Artifacts != nil {
			if m.Runtime.Artifacts == nil {
				m.Runtime.Artifacts = make(map[string]string)
			}
			maps.Copy(m.Runtime.Artifacts, m.Local.Config.Artifacts)
		}
	}

	// 4. Inject Environment Variables passed by scripts (e.g. BOILER_VAR_bl__name)
	// We do this after loading from files so env vars take highest precedence
	for _, envStr := range os.Environ() {
		if strings.HasPrefix(envStr, "BOILER_VAR_") {
			parts := strings.SplitN(envStr, "=", 2)
			if len(parts) == 2 {
				varName := strings.TrimPrefix(parts[0], "BOILER_VAR_")
				if m.Runtime.Vars == nil {
					m.Runtime.Vars = make(map[string]string)
				}
				m.Runtime.Vars[varName] = parts[1]
			}
		}
	}
}

// mergeWithDefaults fills zero-value fields in cfg with values from DefaultConfig.
// This ensures existing users get new fields after a Boiler upgrade.
func mergeWithDefaults(cfg *Config) {
	defaults := DefaultConfig()

	if cfg.DefaultEditor == "" {
		cfg.DefaultEditor = defaults.DefaultEditor
	}
	if cfg.Registry == "" {
		cfg.Registry = defaults.Registry
	}
	if cfg.Name == "" {
		cfg.Name = defaults.Name
	}
	if cfg.Author == "" {
		cfg.Author = defaults.Author
	}
	if cfg.Github == "" {
		cfg.Github = defaults.Github
	}
	if cfg.Version == "" {
		cfg.Version = defaults.Version
	}

	// Ensure paths are populated (in case root changed)
	if cfg.Paths == nil || cfg.Paths.Root == "" {
		cfg.Paths = defaults.Paths
	}

	// Only add artifact keys that the user hasn't set.
	// User's existing values are always respected.
	if cfg.Artifacts == nil {
		cfg.Artifacts = defaults.Artifacts
	} else {
		for k, v := range defaults.Artifacts {
			if _, exists := cfg.Artifacts[k]; !exists {
				cfg.Artifacts[k] = v
			}
		}
	}

	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
	}

	if cfg.Vars == nil {
		cfg.Vars = make(map[string]string)
	} else {
		cfg.Vars = utils.NormalizeVarMap(cfg.Vars)
	}
}
