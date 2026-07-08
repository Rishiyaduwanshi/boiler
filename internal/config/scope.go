package config

// Scope defines the operational context for Boiler commands.
type Scope string

const (
	ScopeGlobal Scope = "global"
	ScopeLocal  Scope = "local"
)
