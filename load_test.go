package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	AppName string       `yaml:"app_name"`
	DB      testDBConfig `yaml:"db"`
	Server  testServer   `yaml:"server"`
}

type testDBConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type testServer struct {
	Address string  `yaml:"address"`
	TLS     testTLS `yaml:"tls"`
}

type testTLS struct {
	Enabled bool `yaml:"enabled"`
}

func TestLoad_Success(t *testing.T) {
	path := writeConfigFile(t, `app_name: demo
db:
  host: localhost
  port: 5432
server:
  address: :8080
  tls:
    enabled: true
`)
	t.Setenv(ConfigPath, path)

	cfg, err := Load[testConfig]()
	if err != nil {
		t.Fatalf("Load returned unexpected error: %v", err)
	}

	if cfg.AppName != "demo" {
		t.Fatalf("unexpected app name: %q", cfg.AppName)
	}

	if cfg.DB.Host != "localhost" || cfg.DB.Port != 5432 {
		t.Fatalf("unexpected db config: %+v", cfg.DB)
	}
}

func TestLoad_PathNotSet(t *testing.T) {
	t.Setenv(ConfigPath, "")

	_, err := Load[testConfig]()
	if !errors.Is(err, ErrPathNotSet) {
		t.Fatalf("expected ErrPathNotSet, got: %v", err)
	}
}

func TestLoad_PathNotExist(t *testing.T) {
	t.Setenv(ConfigPath, filepath.Join(t.TempDir(), "missing.yaml"))

	_, err := Load[testConfig]()
	if !errors.Is(err, ErrPathNotExist) {
		t.Fatalf("expected ErrPathNotExist, got: %v", err)
	}
}

func TestLoadWithDefault_UsesDefault(t *testing.T) {
	t.Setenv(ConfigPath, "")
	path := writeConfigFile(t, `app_name: from-default
db:
  host: db
  port: 1000
server:
  address: :9090
  tls:
    enabled: false
`)

	cfg, err := LoadWithDefault[testConfig](path)
	if err != nil {
		t.Fatalf("LoadWithDefault returned unexpected error: %v", err)
	}

	if cfg.AppName != "from-default" {
		t.Fatalf("unexpected app name: %q", cfg.AppName)
	}
}

func TestLoadWithDefault_EnvOverridesDefault(t *testing.T) {
	defaultPath := writeConfigFile(t, `app_name: from-default
db:
  host: default
  port: 1
server:
  address: :1
  tls:
    enabled: false
`)
	envPath := writeConfigFile(t, `app_name: from-env
db:
  host: env
  port: 2
server:
  address: :2
  tls:
    enabled: true
`)
	t.Setenv(ConfigPath, envPath)

	cfg, err := LoadWithDefault[testConfig](defaultPath)
	if err != nil {
		t.Fatalf("LoadWithDefault returned unexpected error: %v", err)
	}

	if cfg.AppName != "from-env" {
		t.Fatalf("env path must have priority, got: %q", cfg.AppName)
	}
}

func TestMustLoad_PanicsOnError(t *testing.T) {
	t.Setenv(ConfigPath, "")

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	_ = MustLoad[testConfig]()
}

func TestSubConfig_FindsDirectField(t *testing.T) {
	cfg := testConfig{
		DB: testDBConfig{Host: "localhost", Port: 5432},
	}

	db, err := SubConfig[testConfig, testDBConfig](cfg)
	if err != nil {
		t.Fatalf("SubConfig returned unexpected error: %v", err)
	}

	if db.Host != "localhost" || db.Port != 5432 {
		t.Fatalf("unexpected sub config: %+v", db)
	}
}

func TestSubConfig_FindsNestedField(t *testing.T) {
	cfg := testConfig{
		Server: testServer{
			TLS: testTLS{Enabled: true},
		},
	}

	tlsCfg, err := SubConfig[testConfig, testTLS](cfg)
	if err != nil {
		t.Fatalf("SubConfig returned unexpected error: %v", err)
	}

	if !tlsCfg.Enabled {
		t.Fatalf("unexpected nested sub config: %+v", tlsCfg)
	}
}

func TestSubConfig_ReturnsPointer(t *testing.T) {
	cfg := testConfig{
		DB: testDBConfig{Host: "localhost", Port: 5432},
	}

	dbPtr, err := SubConfig[testConfig, *testDBConfig](cfg)
	if err != nil {
		t.Fatalf("SubConfig returned unexpected error: %v", err)
	}

	if dbPtr == nil || dbPtr.Host != "localhost" {
		t.Fatalf("unexpected pointer sub config: %+v", dbPtr)
	}
}

func TestSubConfig_NotFound(t *testing.T) {
	type unknown struct {
		Value string
	}

	_, err := SubConfig[testConfig, unknown](testConfig{})
	if !errors.Is(err, ErrSubConfigNotFound) {
		t.Fatalf("expected ErrSubConfigNotFound, got: %v", err)
	}
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	return path
}
