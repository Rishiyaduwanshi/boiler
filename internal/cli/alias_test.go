package cli

import (
	"reflect"
	"testing"

	"github.com/rishiyaduwanshi/boiler/internal/config"
)

func TestSetAliasFromAssignment_NormalizesAndPersists(t *testing.T) {
	setupVarCommandTest(t)

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
	if got := loaded.Aliases["ll"]; got != "ls" {
		t.Fatalf("loaded.Aliases[ll] = %q", got)
	}
}

func TestSetAliasFromAssignment_AllowsCommandTemplate(t *testing.T) {
	setupVarCommandTest(t)

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
	setupVarCommandTest(t)

	if err := setAliasFromAssignment("ls=search"); err != nil {
		t.Fatalf("setAliasFromAssignment: %v", err)
	}

	if got := cfg.Aliases["ls"]; got != "search" {
		t.Fatalf("cfg.Aliases[ls] = %q", got)
	}
}

func TestUnsetAlias_RemovesNormalizedKey(t *testing.T) {
	setupVarCommandTest(t)

	if err := setAliasFromAssignment("ss=search"); err != nil {
		t.Fatalf("setAliasFromAssignment: %v", err)
	}
	if err := unsetAlias("SS"); err != nil {
		t.Fatalf("unsetAlias: %v", err)
	}

	if _, ok := cfg.Aliases["ss"]; ok {
		t.Fatalf("expected ss to be removed")
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := loaded.Aliases["ss"]; ok {
		t.Fatalf("expected ss to be removed from persisted config")
	}
}

func TestSetAliasFromAssignment_AllowsSelfReference(t *testing.T) {
	setupVarCommandTest(t)

	if err := setAliasFromAssignment("quick=quick --help"); err != nil {
		t.Fatalf("setAliasFromAssignment: %v", err)
	}

	if got := cfg.Aliases["quick"]; got != "quick --help" {
		t.Fatalf("cfg.Aliases[quick] = %q", got)
	}
}

func TestExpandFirstCommandAlias_ReplacesFirstToken(t *testing.T) {
	setupVarCommandTest(t)
	cfg.Aliases["ll"] = "ls"

	got := expandFirstCommandAlias([]string{"ll", "--snippets"})
	want := []string{"ls", "--snippets"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expandFirstCommandAlias() = %v, want %v", got, want)
	}
}

func TestExpandFirstCommandAlias_ExpandsTemplateTokens(t *testing.T) {
	setupVarCommandTest(t)
	cfg.Aliases["sexp"] = "search express -r --registry https://github.com/Rishiyaduwanshi/boiler"

	got := expandFirstCommandAlias([]string{"sexp", "--snippets"})
	want := []string{"search", "express", "-r", "--registry", "https://github.com/Rishiyaduwanshi/boiler", "--snippets"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expandFirstCommandAlias() = %v, want %v", got, want)
	}
}

func TestExpandFirstCommandAlias_IgnoresNonAliasTokens(t *testing.T) {
	setupVarCommandTest(t)
	cfg.Aliases["ll"] = "ls"

	for _, input := range [][]string{{"--help"}, {"unknown"}} {
		got := expandFirstCommandAlias(input)
		if !reflect.DeepEqual(got, input) {
			t.Fatalf("expandFirstCommandAlias(%v) = %v, want %v", input, got, input)
		}
	}
}
