package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testConfig struct {
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Debug bool   `json:"debug"`
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json5")

	content := `{
  // server config
  host: "localhost",
  port: 8080,
  debug: true,
}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	var cfg testConfig
	if err := Load(path, &cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "localhost")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want %d", cfg.Port, 8080)
	}
	if !cfg.Debug {
		t.Error("Debug = false, want true")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	var cfg testConfig
	err := Load("/nonexistent/path.json5", &cfg)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "read file") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoadWithDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json5")

	// File only overrides port
	content := `{ port: 9090 }`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	defaults := testConfig{
		Host:  "0.0.0.0",
		Port:  3000,
		Debug: true,
	}

	var cfg testConfig
	if err := LoadWithDefaults(path, &cfg, defaults); err != nil {
		t.Fatalf("LoadWithDefaults failed: %v", err)
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("Host = %q, want %q (from defaults)", cfg.Host, "0.0.0.0")
	}
	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want %d (from file)", cfg.Port, 9090)
	}
	if !cfg.Debug {
		t.Error("Debug = false, want true (from defaults)")
	}
}

type validatedConfig struct {
	Name    string `json:"name" validate:"required"`
	Address string `json:"address" validate:"required"`
	Note    string `json:"note"`
}

func TestValidatePass(t *testing.T) {
	cfg := validatedConfig{
		Name:    "test",
		Address: "localhost",
		Note:    "",
	}
	if err := Validate(&cfg); err != nil {
		t.Errorf("Validate should pass: %v", err)
	}
}

func TestValidateFail(t *testing.T) {
	cfg := validatedConfig{
		Name: "test",
		// Address is missing
	}
	err := Validate(&cfg)
	if err == nil {
		t.Fatal("expected validation error for missing required field")
	}
	if !strings.Contains(err.Error(), "Address") {
		t.Errorf("error should mention field name, got: %v", err)
	}
}

func TestValidateNonStruct(t *testing.T) {
	s := "not a struct"
	err := Validate(&s)
	if err == nil {
		t.Fatal("expected error for non-struct")
	}
}

func TestValidateByValue(t *testing.T) {
	cfg := validatedConfig{
		Name:    "ok",
		Address: "ok",
	}
	// Pass by value (not pointer)
	if err := Validate(cfg); err != nil {
		t.Errorf("Validate should accept value: %v", err)
	}
}
