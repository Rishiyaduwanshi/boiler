package varcmd

import (
	"testing"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

func setupVarCommandTest(t *testing.T) {
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
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	log, err := utils.NewLogger(cfg.Paths.Logs, false)
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
	if got := loaded.Vars["api_url"]; got != "https://api.example.com" {
		t.Fatalf("loaded.Vars[api_url] = %q", got)
	}
}
