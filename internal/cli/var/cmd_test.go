package varcmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

func setupVarCommandTest(t *testing.T) {
	t.Helper()
	setupScopedVarTest(t, config.ScopeGlobal,
		map[string]string{},
		map[string]string{},
	)
}

func setupScopedVarTest(t *testing.T, scope config.Scope, globalVars, localVars map[string]string) {
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
		forceOverwrite = false
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
}

func TestSetVarFromAssignment_NormalizesAndPersists(t *testing.T) {
	setupVarCommandTest(t)

	if err := setVarFromAssignment("bl__API_URL=https://api.example.com"); err != nil {
		t.Fatalf("setVarFromAssignment: %v", err)
	}

	if got := cfg.Vars["api_url"]; got != "https://api.example.com" {
		t.Fatalf("cfg.Vars[api_url] = %q", got)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := loaded.Runtime.Vars["api_url"]; got != "https://api.example.com" {
		t.Fatalf("loaded.Runtime.Vars[api_url] = %q", got)
	}
}

func TestSetVarFromAssignment_ScopeIsolation_LocalNotPollutingGlobal(t *testing.T) {
	setupScopedVarTest(t, config.ScopeLocal,
		map[string]string{"global_var": "x"},
		map[string]string{},
	)

	if err := setVarFromAssignment("bl__LOCAL_VAR=local-value"); err != nil {
		t.Fatalf("setVarFromAssignment: %v", err)
	}

	if _, ok := config.Ctx.Manager.Global.Config.Vars["local_var"]; ok {
		t.Error("local_var must NOT appear in global config")
	}
	if v := config.Ctx.Manager.Local.Config.Vars["local_var"]; v != "local-value" {
		t.Errorf("local_var should be in local config, got %q", v)
	}
}

func TestSetVarFromAssignment_ExistsWithoutForce(t *testing.T) {
	setupScopedVarTest(t, config.ScopeGlobal,
		map[string]string{"api_url": "https://old.example.com"},
		map[string]string{},
	)
	forceOverwrite = false

	if err := setVarFromAssignment("bl__API_URL=https://new.example.com"); err == nil {
		t.Fatal("expected error when overwriting existing var without --force, got nil")
	}

	if v := config.Ctx.Manager.Global.Config.Vars["api_url"]; v != "https://old.example.com" {
		t.Errorf("original value should be unchanged, got %q", v)
	}
}

func TestSetVarFromAssignment_ExistsWithForce(t *testing.T) {
	setupScopedVarTest(t, config.ScopeGlobal,
		map[string]string{"api_url": "https://old.example.com"},
		map[string]string{},
	)
	forceOverwrite = true

	if err := setVarFromAssignment("bl__API_URL=https://new.example.com"); err != nil {
		t.Fatalf("setVarFromAssignment with --force: %v", err)
	}

	if v := config.Ctx.Manager.Global.Config.Vars["api_url"]; v != "https://new.example.com" {
		t.Errorf("expected 'api_url' to be updated, got %q", v)
	}
}
