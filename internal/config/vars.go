package config

import (
	"fmt"
	"strings"
)

// ScopedVarMap returns vars owned by the active scope's config file, not the merged runtime.
func ScopedVarMap() map[string]string {
	if Ctx.Manager == nil {
		return nil
	}
	if Ctx.Scope == ScopeLocal {
		if Ctx.Manager.Local == nil || Ctx.Manager.Local.Config == nil {
			return nil
		}
		return Ctx.Manager.Local.Config.Vars
	}
	if Ctx.Manager.Global == nil || Ctx.Manager.Global.Config == nil {
		return nil
	}
	return Ctx.Manager.Global.Config.Vars
}

// ScopedVarKey returns the actual stored key in the active scope for a normalized var name.
// Uses case-insensitive matching to handle config files edited via bl conf.
// Returns ("", false) if not found.
func ScopedVarKey(normalizedKey string) (string, bool) {
	m := ScopedVarMap()
	if m == nil {
		return "", false
	}
	if _, ok := m[normalizedKey]; ok {
		return normalizedKey, true
	}
	for k := range m {
		if strings.ToLower(k) == normalizedKey {
			return k, true
		}
	}
	return "", false
}

// SetScopedVar writes key→value into the active scope's config object only.
func SetScopedVar(key, value string) error {
	if Ctx.Manager == nil {
		return fmt.Errorf("config manager not initialized")
	}

	// Remove any existing case-insensitive match before writing the normalized key.
	if storedKey, ok := ScopedVarKey(key); ok {
		_ = DeleteScopedVar(storedKey)
	}

	if Ctx.Scope == ScopeLocal {
		if Ctx.Manager.Local == nil {
			return fmt.Errorf("local config not initialized")
		}
		if Ctx.Manager.Local.Config == nil {
			Ctx.Manager.Local.Config = &Config{}
		}
		if Ctx.Manager.Local.Config.Vars == nil {
			Ctx.Manager.Local.Config.Vars = make(map[string]string)
		}
		Ctx.Manager.Local.Config.Vars[key] = value
		return nil
	}
	if Ctx.Manager.Global == nil {
		return fmt.Errorf("global config not initialized")
	}
	if Ctx.Manager.Global.Config == nil {
		Ctx.Manager.Global.Config = &Config{}
	}
	if Ctx.Manager.Global.Config.Vars == nil {
		Ctx.Manager.Global.Config.Vars = make(map[string]string)
	}
	Ctx.Manager.Global.Config.Vars[key] = value
	return nil
}

// DeleteScopedVar removes key from the active scope's config object only.
func DeleteScopedVar(key string) error {
	if Ctx.Manager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	if Ctx.Scope == ScopeLocal {
		if Ctx.Manager.Local != nil && Ctx.Manager.Local.Config != nil {
			delete(Ctx.Manager.Local.Config.Vars, key)
		}
		return nil
	}
	if Ctx.Manager.Global != nil && Ctx.Manager.Global.Config != nil {
		delete(Ctx.Manager.Global.Config.Vars, key)
	}
	return nil
}

// PersistScopedVars saves the active scope's config file to disk.
func PersistScopedVars() error {
	if Ctx.Manager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	if Ctx.Scope == ScopeLocal {
		if Ctx.Manager.Local == nil {
			return fmt.Errorf("local config not initialized")
		}
		return Ctx.Manager.SaveLocal()
	}
	if Ctx.Manager.Global == nil {
		return fmt.Errorf("global config not initialized")
	}
	return Ctx.Manager.SaveGlobal()
}
