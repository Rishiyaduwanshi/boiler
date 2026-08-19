package unalias

import (
	"os"
	"path/filepath"
	"testing"

	aliascmd "github.com/rishiyaduwanshi/boiler/internal/cli/alias"
	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

// setup creates an isolated test environment with separate Global, Local, and Runtime config objects.
// globalAliases and localAliases seed each scope independently.
func setup(t *testing.T, scope config.Scope, globalAliases, localAliases map[string]string) {
	t.Helper()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	prevCfg := cfg
	prevLogger := logger
	prevCtx := config.Ctx
	t.Cleanup(func() {
		cfg = prevCfg
		logger = prevLogger
		config.Ctx = prevCtx
	})

	merged := make(map[string]string)
	for k, v := range globalAliases {
		merged[k] = v
	}
	for k, v := range localAliases {
		merged[k] = v
	}

	globalCfg := &config.Config{Aliases: globalAliases}
	localCfg := &config.Config{Aliases: localAliases, Scope: string(config.ScopeLocal)}
	runtimeCfg := config.DefaultConfig()
	runtimeCfg.Aliases = merged

	globalPath := filepath.Join(tmp, ".boiler", "boiler.conf.json")
	localPath := filepath.Join(tmp, "project", "boiler.local.json")
	os.MkdirAll(filepath.Dir(globalPath), 0755)
	os.MkdirAll(filepath.Dir(localPath), 0755)
	os.MkdirAll(filepath.Join(tmp, "logs"), 0755)

	cfg = runtimeCfg
	config.Ctx = &config.BoilerContext{
		Manager: &config.Manager{
			Global:  &config.ConfigFile{Path: globalPath, Config: globalCfg, Exists: true},
			Local:   &config.ConfigFile{Path: localPath, Config: localCfg, Exists: true},
			Runtime: runtimeCfg,
		},
		Scope: scope,
	}

	log, err := utils.NewLogger(filepath.Join(tmp, "logs"), false)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	logger = log
	aliascmd.Setup(runtimeCfg, log)
}

// TestUnsetAlias_RemovesFromGlobalScope verifies that a global alias is correctly deleted
// from the global config and persisted to disk.
func TestUnsetAlias_RemovesFromGlobalScope(t *testing.T) {
	setup(t, config.ScopeGlobal,
		map[string]string{"gi": "add github/gitignore:bl__1.gitignore . -m .gitignore -r"},
		map[string]string{},
	)

	if err := unsetAlias("gi"); err != nil {
		t.Fatalf("unsetAlias: %v", err)
	}

	if _, ok := config.Ctx.Manager.Global.Config.Aliases["gi"]; ok {
		t.Error("gi should be removed from global config")
	}
	if _, ok := cfg.Aliases["gi"]; ok {
		t.Error("gi should be removed from runtime map")
	}
}

// TestUnsetAlias_RemovesFromLocalScope verifies that a local alias is correctly deleted
// from the local config without touching global config.
func TestUnsetAlias_RemovesFromLocalScope(t *testing.T) {
	setup(t, config.ScopeLocal,
		map[string]string{"gi": "add github/gitignore:bl__1.gitignore . -m .gitignore -r"},
		map[string]string{"df": "add bl__1 . -m Dockerfile"},
	)

	if err := unsetAlias("df"); err != nil {
		t.Fatalf("unsetAlias: %v", err)
	}

	if _, ok := config.Ctx.Manager.Local.Config.Aliases["df"]; ok {
		t.Error("df should be removed from local config")
	}
	// global must be untouched
	if _, ok := config.Ctx.Manager.Global.Config.Aliases["gi"]; !ok {
		t.Error("gi must remain in global config")
	}
}

// TestUnsetAlias_CrossScopeRejected verifies that attempting to delete a global alias
// while in local scope returns an error and changes nothing.
func TestUnsetAlias_CrossScopeRejected(t *testing.T) {
	setup(t, config.ScopeLocal,
		map[string]string{"gi": "add github/gitignore:bl__1.gitignore . -m .gitignore -r"},
		map[string]string{},
	)

	err := unsetAlias("gi")
	if err == nil {
		t.Fatal("expected error when deleting global alias from local scope, got nil")
	}

	// gi must still exist in global config
	if _, ok := config.Ctx.Manager.Global.Config.Aliases["gi"]; !ok {
		t.Error("gi must remain in global config after failed cross-scope delete")
	}
}

// TestUnsetAlias_CrossScopeRejected_Inverse verifies the same rejection in the opposite direction:
// deleting a local-only alias while in global scope.
func TestUnsetAlias_CrossScopeRejected_Inverse(t *testing.T) {
	setup(t, config.ScopeGlobal,
		map[string]string{},
		map[string]string{"df": "add bl__1 . -m Dockerfile"},
	)

	err := unsetAlias("df")
	if err == nil {
		t.Fatal("expected error when deleting local alias from global scope, got nil")
	}

	if _, ok := config.Ctx.Manager.Local.Config.Aliases["df"]; !ok {
		t.Error("df must remain in local config after failed cross-scope delete")
	}
}

// TestUnsetAlias_NotFound verifies that deleting a completely non-existent alias returns an error.
func TestUnsetAlias_NotFound(t *testing.T) {
	setup(t, config.ScopeGlobal,
		map[string]string{},
		map[string]string{},
	)

	if err := unsetAlias("ghost"); err == nil {
		t.Error("expected error for non-existent alias, got nil")
	}
}

// TestUnsetAlias_InvalidName verifies that an invalid alias name is rejected before any lookup.
func TestUnsetAlias_InvalidName(t *testing.T) {
	setup(t, config.ScopeGlobal,
		map[string]string{},
		map[string]string{},
	)

	if err := unsetAlias("INVALID NAME!"); err == nil {
		t.Error("expected error for invalid alias name, got nil")
	}
}

// TestUnsetAlias_PersistsToDisk verifies that deletion is actually written to disk,
// not just removed from the in-memory runtime map.
func TestUnsetAlias_PersistsToDisk(t *testing.T) {
	setup(t, config.ScopeGlobal,
		map[string]string{"ll": "ls"},
		map[string]string{},
	)

	if err := unsetAlias("ll"); err != nil {
		t.Fatalf("unsetAlias: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := loaded.Runtime.Aliases["ll"]; ok {
		t.Error("ll should not appear after reload — deletion was not persisted")
	}
}

// TestUnsetAlias_MixedCaseKeyInConfig verifies that an alias stored with an uppercase key
// via hand-editing (e.g. bl conf) can still be removed by the normalized name.
func TestUnsetAlias_MixedCaseKeyInConfig(t *testing.T) {
	setup(t, config.ScopeGlobal,
		map[string]string{"LL": "ls"},
		map[string]string{},
	)

	if err := unsetAlias("LL"); err != nil {
		t.Fatalf("unsetAlias with mixed-case stored key: %v", err)
	}

	if _, ok := config.Ctx.Manager.Global.Config.Aliases["LL"]; ok {
		t.Error("LL should be removed from global config")
	}
}
