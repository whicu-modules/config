package config

import (
	"errors"
	"fmt"
)

var (
	ErrPathNotSet   = errors.New("environment variable is not set")
	ErrPathNotExist = errors.New("does not exist")

	ErrSubConfigSourceInvalid = errors.New("source config is invalid")
	ErrSubConfigTargetInvalid = errors.New("target type is invalid")
	ErrSubConfigNotFound      = errors.New("sub config not found")
	ErrSubConfigTypeMismatch  = errors.New("sub config type mismatch")
)

func wrapConfigError(err error) error {
	const prefix = "config"
	return fmt.Errorf("%s: %w", prefix, err)
}
