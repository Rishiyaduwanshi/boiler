package unvar

import (
	"os"
	"path/filepath"
	"testing"

	varcmd "github.com/rishiyaduwanshi/boiler/internal/cli/var"
	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

func setup(t *testing.T, scope config.Scope, globalVars, localVars map[string]string) {
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
	for k, v := range globalVars {
		merged[k] = v
	}
	for k, v := range localVars {
		merged[k] = v
	}

	globalCfg := &config.Config{Vars: globalVars}
	localCfg := &config.Config{Vars: localVars, Scope: string(config.ScopeLocal)}
	runtimeCfg := config.DefaultConfig()
	runtimeCfg.Vars = merged

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
	varcmd.Setup(runtimeCfg, log)
}

func TestUnsetVar_RemovesFromGlobalScope(t *testing.T) {
	setup(t, config.ScopeGlobal,
		map[string]string{"api_url": "https://api.example.com"},
		map[string]string{},
	)

	if err := unsetVar("bl__API_URL"); err != nil {
		t.Fatalf("unsetVar: %v", err)
	}

	if _, ok := config.Ctx.Manager.Global.Config.Vars["api_url"]; ok {
		t.Error("api_url should be removed from global config")
	}
}

func TestUnsetVar_RemovesFromLocalScope(t *testing.T) {
	setup(t, config.ScopeLocal,
		map[string]string{"global_var": "x"},
		map[string]string{"local_var": "y"},
	)

	if err := unsetVar("bl__LOCAL_VAR"); err != nil {
		t.Fatalf("unsetVar: %v", err)
	}

	if _, ok := config.Ctx.Manager.Local.Config.Vars["local_var"]; ok {
		t.Error("local_var should be removed from local config")
	}
	if _, ok := config.Ctx.Manager.Global.Config.Vars["global_var"]; !ok {
		t.Error("global_var must remain in global config")
	}
}

func TestUnsetVar_CrossScopeRejected(t *testing.T) {
	setup(t, config.ScopeLocal,
		map[string]string{"global_var": "x"},
		map[string]string{},
	)

	if err := unsetVar("bl__GLOBAL_VAR"); err == nil {
		t.Fatal("expected error when deleting global var from local scope, got nil")
	}

	if _, ok := config.Ctx.Manager.Global.Config.Vars["global_var"]; !ok {
		t.Error("global_var must remain in global config after failed cross-scope delete")
	}
}

func TestUnsetVar_CrossScopeRejected_Inverse(t *testing.T) {
	setup(t, config.ScopeGlobal,
		map[string]string{},
		map[string]string{"local_var": "y"},
	)

	if err := unsetVar("bl__LOCAL_VAR"); err == nil {
		t.Fatal("expected error when deleting local var from global scope, got nil")
	}

	if _, ok := config.Ctx.Manager.Local.Config.Vars["local_var"]; !ok {
		t.Error("local_var must remain in local config after failed cross-scope delete")
	}
}

func TestUnsetVar_NotFound(t *testing.T) {
	setup(t, config.ScopeGlobal,
		map[string]string{},
		map[string]string{},
	)

	if err := unsetVar("bl__GHOST"); err == nil {
		t.Error("expected error for non-existent var, got nil")
	}
}
