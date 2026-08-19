package config

import "fmt"

// ScopedAliasMap returns aliases owned by the active scope's config file, not the merged runtime.
func ScopedAliasMap() map[string]string {
	if Ctx.Manager == nil {
		return nil
	}
	if Ctx.Scope == ScopeLocal {
		if Ctx.Manager.Local == nil || Ctx.Manager.Local.Config == nil {
			return nil
		}
		return Ctx.Manager.Local.Config.Aliases
	}
	if Ctx.Manager.Global == nil || Ctx.Manager.Global.Config == nil {
		return nil
	}
	return Ctx.Manager.Global.Config.Aliases
}

// SetScopedAlias writes name→target into the active scope's config object only.
func SetScopedAlias(name, target string) error {
	if Ctx.Manager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	if Ctx.Scope == ScopeLocal {
		if Ctx.Manager.Local == nil {
			return fmt.Errorf("local config not initialized")
		}
		if Ctx.Manager.Local.Config == nil {
			Ctx.Manager.Local.Config = &Config{}
		}
		if Ctx.Manager.Local.Config.Aliases == nil {
			Ctx.Manager.Local.Config.Aliases = make(map[string]string)
		}
		Ctx.Manager.Local.Config.Aliases[name] = target
		return nil
	}
	if Ctx.Manager.Global == nil {
		return fmt.Errorf("global config not initialized")
	}
	if Ctx.Manager.Global.Config == nil {
		Ctx.Manager.Global.Config = &Config{}
	}
	if Ctx.Manager.Global.Config.Aliases == nil {
		Ctx.Manager.Global.Config.Aliases = make(map[string]string)
	}
	Ctx.Manager.Global.Config.Aliases[name] = target
	return nil
}

// DeleteScopedAlias removes name from the active scope's config object only.
func DeleteScopedAlias(name string) error {
	if Ctx.Manager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	if Ctx.Scope == ScopeLocal {
		if Ctx.Manager.Local != nil && Ctx.Manager.Local.Config != nil {
			delete(Ctx.Manager.Local.Config.Aliases, name)
		}
		return nil
	}
	if Ctx.Manager.Global != nil && Ctx.Manager.Global.Config != nil {
		delete(Ctx.Manager.Global.Config.Aliases, name)
	}
	return nil
}

// PersistScopedAliases saves the active scope's config file to disk.
func PersistScopedAliases() error {
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
