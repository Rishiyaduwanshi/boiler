package alias

import (
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
