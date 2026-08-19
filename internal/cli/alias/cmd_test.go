package alias

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

func setupAliasCommandTest(t *testing.T) {
	t.Helper()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	previousCfg := cfg
	previousLogger := logger
	t.Cleanup(func() {
		cfg = previousCfg
		logger = previousLogger
	})

	cfg = config.DefaultConfig()
	if err := cfg.InitializeDirs(); err != nil {
		t.Fatalf("InitializeDirs: %v", err)
	}

	globalPath, _ := config.GlobalConfigPath()
	config.Ctx = &config.BoilerContext{
		Manager: &config.Manager{
			Global:  &config.ConfigFile{Path: globalPath, Config: cfg},
			Runtime: cfg,
		},
		Scope: config.ScopeGlobal,
	}

	log, err := utils.NewLogger(cfg.Paths.Logs, false)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	logger = log
}

func TestSetAliasFromAssignment_NormalizesAndPersists(t *testing.T) {
	setupAliasCommandTest(t)

	if err := setAliasFromAssignment("LL=ls"); err != nil {
		t.Fatalf("setAliasFromAssignment: %v", err)
	}

	if got := cfg.Aliases["ll"]; got != "ls" {
		t.Fatalf("cfg.Aliases[ll] = %q", got)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := loaded.Runtime.Aliases["ll"]; got != "ls" {
		t.Fatalf("loaded.Runtime.Aliases[ll] = %q", got)
	}
}

func TestSetAliasFromAssignment_AllowsCommandTemplate(t *testing.T) {
	setupAliasCommandTest(t)

	assignment := "sexp=bl search express -r --registry https://github.com/Rishiyaduwanshi/boiler"
	if err := setAliasFromAssignment(assignment); err != nil {
		t.Fatalf("setAliasFromAssignment: %v", err)
	}

	want := "search express -r --registry https://github.com/Rishiyaduwanshi/boiler"
	if got := cfg.Aliases["sexp"]; got != want {
		t.Fatalf("cfg.Aliases[sexp] = %q, want %q", got, want)
	}
}

func TestSetAliasFromAssignment_AllowsBuiltInCommandName(t *testing.T) {
	setupAliasCommandTest(t)

	if err := setAliasFromAssignment("ls=search"); err != nil {
		t.Fatalf("setAliasFromAssignment: %v", err)
	}

	if got := cfg.Aliases["ls"]; got != "search" {
		t.Fatalf("cfg.Aliases[ls] = %q", got)
	}
}

func TestSetAliasFromAssignment_AllowsSelfReference(t *testing.T) {
	setupAliasCommandTest(t)

	if err := setAliasFromAssignment("quick=quick --help"); err != nil {
		t.Fatalf("setAliasFromAssignment: %v", err)
	}

	if got := cfg.Aliases["quick"]; got != "quick --help" {
		t.Fatalf("cfg.Aliases[quick] = %q", got)
	}
}

func TestExpandFirstCommandAlias_ReplacesFirstToken(t *testing.T) {
	setupAliasCommandTest(t)
	cfg.Aliases["ll"] = "ls"

	got := ExpandFirstCommandAlias([]string{"ll", "--snippets"})
	want := []string{"ls", "--snippets"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ExpandFirstCommandAlias() = %v, want %v", got, want)
	}
}

func TestExpandFirstCommandAlias_ExpandsTemplateTokens(t *testing.T) {
	setupAliasCommandTest(t)
	cfg.Aliases["sexp"] = "search express -r --registry https://github.com/Rishiyaduwanshi/boiler"

	got := ExpandFirstCommandAlias([]string{"sexp", "--snippets"})
	want := []string{"search", "express", "-r", "--registry", "https://github.com/Rishiyaduwanshi/boiler", "--snippets"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ExpandFirstCommandAlias() = %v, want %v", got, want)
	}
}

func TestExpandFirstCommandAlias_IgnoresNonAliasTokens(t *testing.T) {
	setupAliasCommandTest(t)
	cfg.Aliases["ll"] = "ls"

	for _, input := range [][]string{{"--help"}, {"unknown"}} {
		got := ExpandFirstCommandAlias(input)
		if !reflect.DeepEqual(got, input) {
			t.Fatalf("ExpandFirstCommandAlias(%v) = %v, want %v", input, got, input)
		}
	}
}

// setupScopedAliasTest builds a test environment where Global.Config, Local.Config, and Runtime
// are SEPARATE objects — the real-world scenario that the existing setupAliasCommandTest misses
// by reusing the same pointer for all three.
func setupScopedAliasTest(t *testing.T, scope config.Scope) {
	t.Helper()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	previousCfg := cfg
	previousLogger := logger
	previousCtx := config.Ctx
	t.Cleanup(func() {
		cfg = previousCfg
		logger = previousLogger
		config.Ctx = previousCtx
	})

	// Separate objects — global-only and local-only aliases must not bleed between files.
	globalCfg := &config.Config{
		Aliases: map[string]string{"gi": "add github/gitignore:bl__1.gitignore . -m .gitignore -r"},
	}
	localCfg := &config.Config{
		Aliases: map[string]string{"df": "add bl__1 . -m Dockerfile"},
		Scope:   string(config.ScopeLocal),
	}
	runtimeCfg := config.DefaultConfig()
	runtimeCfg.Aliases = map[string]string{
		"gi": "add github/gitignore:bl__1.gitignore . -m .gitignore -r",
		"df": "add bl__1 . -m Dockerfile",
	}

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

// TestSetAliasFromAssignment_GlobalScopeDoesNotPollutLocalConfig verifies that setting
// an alias in global scope does not write it (or any global alias) into boiler.local.json.
func TestSetAliasFromAssignment_GlobalScopeDoesNotPollutLocalConfig(t *testing.T) {
	setupScopedAliasTest(t, config.ScopeGlobal)

	if err := setAliasFromAssignment("newcmd=search --all"); err != nil {
		t.Fatalf("setAliasFromAssignment: %v", err)
	}

	// New alias must be in global config.
	if _, ok := config.Ctx.Manager.Global.Config.Aliases["newcmd"]; !ok {
		t.Error("global config should contain newcmd alias")
	}
	// Local config must NOT gain the new alias.
	if _, ok := config.Ctx.Manager.Local.Config.Aliases["newcmd"]; ok {
		t.Error("local config must NOT contain newcmd — scope isolation broken")
	}
	// Global config must NOT absorb df (the local-only alias from the merged Runtime).
	if _, ok := config.Ctx.Manager.Global.Config.Aliases["df"]; ok {
		t.Error("global config must NOT contain df — local alias leaked into global file")
	}
}

// TestSetAliasFromAssignment_LocalScopeDoesNotPollutGlobalConfig verifies that setting
// an alias in local scope does not write it (or any local alias) into boiler.conf.json.
func TestSetAliasFromAssignment_LocalScopeDoesNotPollutGlobalConfig(t *testing.T) {
	setupScopedAliasTest(t, config.ScopeLocal)

	if err := setAliasFromAssignment("localcmd=add bl__1 ./src"); err != nil {
		t.Fatalf("setAliasFromAssignment: %v", err)
	}

	// New alias must be in local config.
	if _, ok := config.Ctx.Manager.Local.Config.Aliases["localcmd"]; !ok {
		t.Error("local config should contain localcmd alias")
	}
	// Global config must NOT gain the new alias.
	if _, ok := config.Ctx.Manager.Global.Config.Aliases["localcmd"]; ok {
		t.Error("global config must NOT contain localcmd — scope isolation broken")
	}
	// Local config must NOT absorb gi (the global-only alias from the merged Runtime).
	if _, ok := config.Ctx.Manager.Local.Config.Aliases["gi"]; ok {
		t.Error("local config must NOT contain gi — global alias leaked into local file")
	}
}

// TestDeleteScopedAlias_RemovesOnlyFromActiveScope verifies that unaliasing in local scope
// does not touch the global config file.
func TestDeleteScopedAlias_RemovesOnlyFromActiveScope(t *testing.T) {
	setupScopedAliasTest(t, config.ScopeLocal)

	if err := config.DeleteScopedAlias("df"); err != nil {
		t.Fatalf("DeleteScopedAlias: %v", err)
	}

	// df must be gone from local config.
	if _, ok := config.Ctx.Manager.Local.Config.Aliases["df"]; ok {
		t.Error("local config should no longer contain df after delete")
	}
	// gi must remain untouched in global config.
	if _, ok := config.Ctx.Manager.Global.Config.Aliases["gi"]; !ok {
		t.Error("global config should still contain gi — unrelated alias deleted")
	}
}

// TestPersistConfigAliases_NilManagerReturnsError ensures a nil Manager produces
// a clean error instead of a nil-pointer panic.
func TestPersistConfigAliases_NilManagerReturnsError(t *testing.T) {
	prev := config.Ctx
	config.Ctx = &config.BoilerContext{Manager: nil, Scope: config.ScopeGlobal}
	t.Cleanup(func() { config.Ctx = prev })

	if err := config.PersistScopedAliases(); err == nil {
		t.Error("expected error for nil Manager, got nil")
	}
}

// TestScopedAliasMap_GlobalScopeReturnsOnlyGlobalAliases verifies that
// ScopedAliasMap returns global-owned aliases only, not local ones.
func TestScopedAliasMap_GlobalScopeReturnsOnlyGlobalAliases(t *testing.T) {
	setupScopedAliasTest(t, config.ScopeGlobal)

	got := config.ScopedAliasMap()

	if _, ok := got["gi"]; !ok {
		t.Error("ScopedAliasMap should contain gi (global alias)")
	}
	if _, ok := got["df"]; ok {
		t.Error("ScopedAliasMap must NOT contain df (local-only alias)")
	}
}

// TestScopedAliasMap_LocalScopeReturnsOnlyLocalAliases verifies that
// ScopedAliasMap returns local-owned aliases only, not global ones.
func TestScopedAliasMap_LocalScopeReturnsOnlyLocalAliases(t *testing.T) {
	setupScopedAliasTest(t, config.ScopeLocal)

	got := config.ScopedAliasMap()

	if _, ok := got["df"]; !ok {
		t.Error("ScopedAliasMap should contain df (local alias)")
	}
	if _, ok := got["gi"]; ok {
		t.Error("ScopedAliasMap must NOT contain gi (global-only alias)")
	}
}

// TestScopedAliasMap_NilManagerReturnsNil verifies nil-safety.
func TestScopedAliasMap_NilManagerReturnsNil(t *testing.T) {
	prev := config.Ctx
	config.Ctx = &config.BoilerContext{Manager: nil, Scope: config.ScopeGlobal}
	t.Cleanup(func() { config.Ctx = prev })

	if got := config.ScopedAliasMap(); got != nil {
		t.Errorf("expected nil for nil Manager, got %v", got)
	}
}
