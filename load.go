package config

import (
	"cmp"
	"errors"
	"os"
	"reflect"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	ConfigPath = "PATH_CONFIG"
)

func Load[T any]() (t T, err error) {
	defer func() {
		if err != nil {
			err = wrapConfigError(err)
		}
	}()
	var zero T
	path, err := resolvePath("")
	if err != nil {
		return zero, err
	}

	return readConfig[T](path)
}

func MustLoad[T any]() T {
	cfg, err := Load[T]()
	if err != nil {
		panic(err)
	}
	return cfg
}

func LoadWithDefault[T any](def string) (t T, err error) {
	defer func() {
		if err != nil {
			err = wrapConfigError(err)
		}
	}()
	var zero T
	path, err := resolvePath(def)
	if err != nil {
		return zero, err
	}

	return readConfig[T](path)
}

func MustLoadWithDefault[T any](def string) T {
	cfg, err := LoadWithDefault[T](def)
	if err != nil {
		panic(err)
	}
	return cfg
}

func SubConfig[T any, S any](cfg T) (s S, err error) {
	defer func() {
		if err != nil {
			err = wrapConfigError(err)
		}
	}()

	targetType := reflect.TypeFor[S]()

	value := reflect.ValueOf(cfg)
	if !value.IsValid() {
		return s, ErrSubConfigSourceInvalid
	}

	found, ok := findSubConfigValue(value, targetType)
	if !ok {
		return s, ErrSubConfigNotFound
	}

	out, ok := toTargetType(found, targetType)
	if !ok {
		return s, ErrSubConfigTypeMismatch
	}

	return out.Interface().(S), nil
}

func resolvePath(def string) (string, error) {
	path := cmp.Or(os.Getenv(ConfigPath), def)
	if path == "" {
		return "", ErrPathNotSet
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return "", ErrPathNotExist
	}

	return path, nil
}

func readConfig[T any](path string) (T, error) {
	var cfg T
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func findSubConfigValue(value reflect.Value, targetType reflect.Type) (reflect.Value, bool) {
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, false
		}
		value = value.Elem()
	}

	if !value.IsValid() {
		return reflect.Value{}, false
	}

	if value.Kind() == reflect.Struct && !value.CanAddr() {
		addressable := reflect.New(value.Type()).Elem()
		addressable.Set(value)
		value = addressable
	}

	if out, ok := toTargetType(value, targetType); ok {
		return out, true
	}

	if value.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	for i := range value.NumField() {
		if field := value.Field(i); field.IsValid() {
			if found, ok := findSubConfigValue(field, targetType); ok {
				return found, true
			}
		}
	}

	return reflect.Value{}, false
}

func toTargetType(value reflect.Value, targetType reflect.Type) (reflect.Value, bool) {
	if !value.IsValid() {
		return reflect.Value{}, false
	}

	if value.Type().ConvertibleTo(targetType) {
		return value.Convert(targetType), true
	}

	if targetType.Kind() == reflect.Pointer && value.CanAddr() && value.Addr().Type().ConvertibleTo(targetType) {
		return value.Addr().Convert(targetType), true
	}

	if value.Kind() == reflect.Pointer && !value.IsNil() && value.Elem().Type().ConvertibleTo(targetType) {
		return value.Elem().Convert(targetType), true
	}

	return reflect.Value{}, false
}
