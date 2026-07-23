// SPDX-License-Identifier: GPL-3.0-or-later
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	JsonTree           bool   `mapstructure:"jsonTree"`
	TxtTree            bool   `mapstructure:"txtTree"`
	TerminalTree       bool   `mapstructure:"terminalTree"`
	AnnotateTree       bool   `mapstructure:"annotateTree"`
	Density            int    `mapstructure:"density"`
	AnnotationsPadding int    `mapstructure:"annotationsPadding"`
	FilesFirst         bool   `mapstructure:"filesFirst"`
	HiddenFiles        bool   `mapstructure:"hiddenFiles"`
	Alphabetic         bool   `mapstructure:"alphabetic"`
	TreeSet            int    `mapstructure:"connectorSet"`
	MaxDepth           int    `mapstructure:"maxDepth"`
	MaxElements        int    `mapstructure:"maxElements"`
	DirHints           bool   `mapstructure:"dirHints"`
	PowerLevel         string `mapstructure:"powerLevel"`
}

func loadConfig(programName, defaultConfig string) (*Config, error) {
	configPath := getConfigPath(programName)
	configFile := filepath.Join(configPath, "config.yaml")

	// Create default config if it doesn't exist
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := createDefaultConfig(programName, defaultConfig); err != nil {
			return nil, fmt.Errorf("error creating default config: %w", err)
		}
		infoLog.Printf("Created default config at: %s\n", configFile)
		infoLog.Println("To edit the configuration, set a flag's value and add the -saveconfig flag")
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")

	viper.SetDefault("jsonTree", false)
	viper.SetDefault("txtTree", false)
	viper.SetDefault("terminalTree", true)
	viper.SetDefault("annotateTree", false)
	viper.SetDefault("density", 3)
	viper.SetDefault("annotationsPadding", 3)
	viper.SetDefault("filesFirst", true)
	viper.SetDefault("hiddenFiles", false)
	viper.SetDefault("alphabetic", true)
	viper.SetDefault("connectorSet", 2)
	viper.SetDefault("maxDepth", 10)
	viper.SetDefault("maxElements", 20)
	viper.SetDefault("dirHints", true)
	viper.SetDefault("powerLevel", "m")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func saveConfig(config *Config, programName string) error {
	configPath := getConfigPath(programName)
	configFile := filepath.Join(configPath, "config.yaml")

	viper.Set("jsonTree", config.JsonTree)
	viper.Set("txtTree", config.TxtTree)
	viper.Set("terminalTree", config.TerminalTree)
	viper.Set("annotateTree", config.AnnotateTree)
	viper.Set("density", config.Density)
	viper.Set("annotationsPadding", config.AnnotationsPadding)
	viper.Set("filesFirst", config.FilesFirst)
	viper.Set("hiddenFiles", config.HiddenFiles)
	viper.Set("alphabetic", config.Alphabetic)
	viper.Set("connectorSet", config.TreeSet)
	viper.Set("maxDepth", config.MaxDepth)
	viper.Set("maxElements", config.MaxElements)
	viper.Set("dirHints", config.DirHints)
	viper.Set("powerLevel", config.PowerLevel)

	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

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

func createDefaultConfig(programName, defaultConfig string) error {
	configPath := getConfigPath(programName)

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	configFile := filepath.Join(configPath, "config.yaml")

	return os.WriteFile(configFile, []byte(defaultConfig), 0644)
}
