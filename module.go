package config

import (
	"go.uber.org/fx"
)

func NewConfigModule[T any](moduleName string) fx.Option {
	return fx.Module(moduleName, fx.Provide(Load[T]))
}

func NewConfig[T any]() fx.Option {
	return fx.Provide(Load[T])
}

func NewSubConfigModule[T any, S any](moduleName string) fx.Option {
	return fx.Module(moduleName, fx.Provide(SubConfig[T, S]))
}

func NewSubConfig[T any, S any]() fx.Option {
	return fx.Provide(SubConfig[T, S])
}
