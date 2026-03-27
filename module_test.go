package config

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/fx"
)

const testConfigYAML = `app_name: module
db:
  host: db-host
  port: 15432
server:
  address: :8080
  tls:
    enabled: false
`

const testConfigYAMLWithDifferentDB = `app_name: module
db:
  host: localhost
  port: 5432
server:
  address: :8080
  tls:
    enabled: true
`

func TestNewConfigModule_ProvidesConfig(t *testing.T) {
	path := writeConfigFile(t, testConfigYAMLWithDifferentDB)
	t.Setenv(ConfigPath, path)

	var got testConfig
	app := fx.New(
		fx.NopLogger,
		NewConfigModule[testConfig]("config"),
		fx.Invoke(func(cfg testConfig) {
			got = cfg
		}),
	)

	startAndStopApp(t, app)

	if got == (testConfig{}) {
		t.Fatal("expected config to be provided")
	}

	if got.AppName != "module" {
		t.Fatalf("unexpected app name: %q", got.AppName)
	}
}

func TestNewConfigModule_ReturnsErrorOnMissingPath(t *testing.T) {
	t.Setenv(ConfigPath, "")

	app := fx.New(
		fx.NopLogger,
		NewConfigModule[testConfig]("config"),
		fx.Invoke(func(_ testConfig) {}),
	)

	ctx := t.Context()
	err := app.Start(ctx)
	if err == nil {
		t.Fatal("expected start error")
	}

	if !errors.Is(err, ErrPathNotSet) {
		t.Fatalf("expected ErrPathNotSet, got: %v", err)
	}
}

func TestNewSubConfigModule_ProvidesSubConfig(t *testing.T) {
	path := writeConfigFile(t, testConfigYAML)
	t.Setenv(ConfigPath, path)

	var got testDBConfig
	app := fx.New(
		fx.NopLogger,
		NewConfigModule[testConfig]("config"),
		NewSubConfigModule[testConfig, testDBConfig]("db"),
		fx.Invoke(func(db testDBConfig) {
			got = db
		}),
	)

	startAndStopApp(t, app)

	if got.Host != "db-host" || got.Port != 15432 {
		t.Fatalf("unexpected sub config: %+v", got)
	}
}

func TestNewSubConfig_ProvidesSubConfig(t *testing.T) {
	path := writeConfigFile(t, testConfigYAML)
	t.Setenv(ConfigPath, path)

	var got testDBConfig
	app := fx.New(
		fx.NopLogger,
		NewConfigModule[testConfig]("config"),
		NewSubConfig[testConfig, testDBConfig](),
		fx.Invoke(func(db testDBConfig) {
			got = db
		}),
	)

	startAndStopApp(t, app)

	if got.Host != "db-host" || got.Port != 15432 {
		t.Fatalf("unexpected sub config: %+v", got)
	}
}

func TestNewConfig_ProvidesConfig(t *testing.T) {
	path := writeConfigFile(t, testConfigYAMLWithDifferentDB)
	t.Setenv(ConfigPath, path)

	var got testConfig
	app := fx.New(
		fx.NopLogger,
		NewConfig[testConfig](),
		fx.Invoke(func(cfg testConfig) {
			got = cfg
		}),
	)

	startAndStopApp(t, app)

	if got == (testConfig{}) {
		t.Fatal("expected config to be provided")
	}

	if got.DB.Host != "localhost" || got.DB.Port != 5432 {
		t.Fatalf("unexpected config: %+v", got)
	}
}

func TestNewSubConfig_ReturnsErrorWhenSubConfigMissing(t *testing.T) {
	path := writeConfigFile(t, testConfigYAML)
	t.Setenv(ConfigPath, path)

	type unknownSubConfig struct {
		Value string
	}

	app := fx.New(
		fx.NopLogger,
		NewConfig[testConfig](),
		NewSubConfig[testConfig, unknownSubConfig](),
		fx.Invoke(func(_ unknownSubConfig) {}),
	)

	err := app.Start(t.Context())
	if err == nil {
		t.Fatal("expected start error")
	}

	if !errors.Is(err, ErrSubConfigNotFound) {
		t.Fatalf("expected ErrSubConfigNotFound, got: %v", err)
	}
}

func startAndStopApp(t *testing.T, app *fx.App) {
	t.Helper()

	ctx := t.Context()
	if err := app.Start(ctx); err != nil {
		t.Fatalf("app start failed: %v", err)
	}

	t.Cleanup(func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := app.Stop(stopCtx)
		if err != nil {
			t.Fatalf("app stop failed: %v", err)
		}
	})
}
