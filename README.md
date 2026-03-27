# config

A lightweight module for loading configuration from a file path defined in the `PATH_CONFIG` environment variable.

## Installation

```bash
go get github.com/whicu-modules/config
```

## API

- `Load[T any]() (T, error)`
- `MustLoad[T any]() T`
- `LoadWithDefault[T any](def string) (T, error)`
- `MustLoadWithDefault[T any](def string) T`
- `SubConfig[T any, S any](cfg T) (S, error)`
- `NewConfigModule[T any](moduleName string) fx.Option`
- `NewConfig[T any]() fx.Option`
- `NewSubConfigModule[T any, S any](moduleName string) fx.Option`
- `NewSubConfig[T any, S any]() fx.Option`

## Basic Usage

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/whicu-modules/config"
)

type DBConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type AppConfig struct {
	AppName string   `yaml:"app_name"`
	DB      DBConfig `yaml:"db"`
}

func main() {
	os.Setenv("PATH_CONFIG", "./config.yaml")

	cfg, err := config.Load[AppConfig]()
	if err != nil {
		log.Fatal(err)
	}

	db, err := config.SubConfig[AppConfig, DBConfig](cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg.AppName)
	fmt.Println(db.Host, db.Port)
}
```

## Fx Integration

```go
app := fx.New(
	config.NewConfigModule[AppConfig]("config"),
	config.NewSubConfigModule[AppConfig, DBConfig]("db"),
	fx.Invoke(func(cfg AppConfig, db DBConfig) {
		// use cfg and db
	}),
)
```

You can also use non-module providers:

```go
app := fx.New(
	config.NewConfig[AppConfig](),
	config.NewSubConfig[AppConfig, DBConfig](),
)
```

## How SubConfig Works

`SubConfig` searches for the first matching sub-structure by type in the provided config.

- nested structure traversal is supported;
- both value and pointer target types are supported;
- if no matching sub-structure is found, `ErrSubConfigNotFound` is returned.

## Errors

All public functions wrap errors with the `config` prefix.

- `ErrPathNotSet` - environment variable is not set;
- `ErrPathNotExist` - config file does not exist;
- `ErrSubConfigNotFound` - sub-config was not found;
- `ErrSubConfigTypeMismatch` - found value cannot be converted to the requested target type.
