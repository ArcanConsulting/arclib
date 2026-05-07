// Package config provides a configuration file loader using JSON5.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"arcan-it.de/arclib/json5"
)

// Load reads a JSON5 configuration file and unmarshals it into target.
func Load(path string, target interface{}) error {
	data, err := os.ReadFile(path) //nolint:gosec // path is intentionally caller-provided
	if err != nil {
		return fmt.Errorf("config: read file: %w", err)
	}
	if err := json5.Unmarshal(data, target); err != nil {
		return fmt.Errorf("config: parse: %w", err)
	}
	return nil
}

// LoadWithDefaults reads a JSON5 configuration file and unmarshals it into
// target, using defaults as the base values. Fields not present in the file
// retain their default values.
func LoadWithDefaults(path string, target, defaults interface{}) error {
	// Marshal defaults to JSON, then unmarshal into target to set base values.
	defaultBytes, err := json.Marshal(defaults)
	if err != nil {
		return fmt.Errorf("config: marshal defaults: %w", err)
	}
	if err := json.Unmarshal(defaultBytes, target); err != nil {
		return fmt.Errorf("config: apply defaults: %w", err)
	}

	// Read and overlay the config file.
	data, err := os.ReadFile(path) //nolint:gosec // path is intentionally caller-provided
	if err != nil {
		return fmt.Errorf("config: read file: %w", err)
	}
	if err := json5.Unmarshal(data, target); err != nil {
		return fmt.Errorf("config: parse: %w", err)
	}
	return nil
}

// Validate checks that all struct fields tagged with `validate:"required"` are
// non-zero values. It returns an error listing the first missing required field.
func Validate(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("config: validate expects a struct, got %s", val.Kind())
	}

	t := val.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		tag := field.Tag.Get("validate")
		if tag != "required" {
			continue
		}
		if val.Field(i).IsZero() {
			return fmt.Errorf("config: field %q is required but has zero value", field.Name)
		}
	}
	return nil
}
