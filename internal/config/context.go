package config

// Ctx holds the resolved scope and manager for the current CLI execution.
var Ctx *BoilerContext

// ResolveScope determines the current operational scope based on priority:
// 1. CLI Flags (--global or --local)
// 2. Nearest Project Config (`scope` field)
// 3. Global Config (`scope` field)
// 4. Default fallback to ScopeGlobal
func ResolveScope(m *Manager, globalFlag, localFlag bool) Scope {
	// 1. CLI Flags always win
	if globalFlag {
		return ScopeGlobal
	}
	if localFlag {
		return ScopeLocal
	}

	// 2. Nearest Project Config
	if m.Local != nil && m.Local.Exists && m.Local.Config != nil {
		if m.Local.Config.Scope == string(ScopeLocal) || m.Local.Config.Scope == string(ScopeGlobal) {
			return Scope(m.Local.Config.Scope)
		}
	}

	// 3. Global Config
	if m.Global != nil && m.Global.Exists && m.Global.Config != nil {
		if m.Global.Config.Scope == string(ScopeLocal) || m.Global.Config.Scope == string(ScopeGlobal) {
			return Scope(m.Global.Config.Scope)
		}
	}

	// 4. Default
	return ScopeGlobal
}
