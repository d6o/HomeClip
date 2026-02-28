package config

import (
	"os"
	"testing"
)

func TestNewConfig_Defaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("DATA_DIR")

	cfg := NewConfig()

	if cfg.Port != defaultPort {
		t.Errorf("expected port %q, got %q", defaultPort, cfg.Port)
	}
	if cfg.DataDir != defaultDataDir {
		t.Errorf("expected data dir %q, got %q", defaultDataDir, cfg.DataDir)
	}
}

func TestNewConfig_CustomValues(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATA_DIR", "/tmp/custom")

	cfg := NewConfig()

	if cfg.Port != "9090" {
		t.Errorf("expected port %q, got %q", "9090", cfg.Port)
	}
	if cfg.DataDir != "/tmp/custom" {
		t.Errorf("expected data dir %q, got %q", "/tmp/custom", cfg.DataDir)
	}
}

func TestNewConfig_PartialOverride(t *testing.T) {
	t.Setenv("PORT", "3000")
	os.Unsetenv("DATA_DIR")

	cfg := NewConfig()

	if cfg.Port != "3000" {
		t.Errorf("expected port %q, got %q", "3000", cfg.Port)
	}
	if cfg.DataDir != defaultDataDir {
		t.Errorf("expected data dir %q, got %q", defaultDataDir, cfg.DataDir)
	}
}
