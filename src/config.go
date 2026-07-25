// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/spf13/viper"
)

// LoadConfig loads a named yaml config file from configPath into a new T,
// creating the file with the given defaults if it doesn't exist yet.
func LoadConfig[T any](configPath, configName string, defaults map[string]any) (*T, error) {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	v.AddConfigPath(".")

	for key, value := range defaults {
		v.SetDefault(key, value)
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg T
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// SaveConfig writes any mapstructure-tagged struct to a yaml config file.
func SaveConfig(config any, configPath, configName string) error {
	v := viper.New()
	v.SetConfigType("yaml")

	for key, value := range structToMap(config) {
		v.Set(key, value)
	}

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	configFile := filepath.Join(configPath, configName+".yaml")
	if err := v.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// structToMap flattens a struct's fields into a map keyed by mapstructure tag
// (falling back to the field name), so callers don't need to enumerate fields.
func structToMap(cfg any) map[string]any {
	result := make(map[string]any)

	val := reflect.ValueOf(cfg)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		key := field.Tag.Get("mapstructure")
		if key == "" {
			key = field.Name
		}
		result[key] = val.Field(i).Interface()
	}

	return result
}

// getConfigPath resolves the config directory for a given program name,
// preferring XDG_CONFIG_HOME.
func getConfigPath(programName string) string {
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, programName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	return filepath.Join(home, ".config", programName)
}
