// SPDX-License-Identifier: GPL-3.0-or-later
package main

import (
	"log"
	"os"
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

var configDefaults = map[string]any{
	"jsonTree":           false,
	"txtTree":            false,
	"terminalTree":       true,
	"annotateTree":       false,
	"density":            3,
	"annotationsPadding": 3,
	"filesFirst":         true,
	"hiddenFiles":        false,
	"alphabetic":         true,
	"connectorSet":       2,
	"maxDepth":           10,
	"maxElements":        20,
	"dirHints":           true,
	"powerLevel":         "m",
}

// loadConfig and saveConfig are the app-specific entry points used by
// main_cli.go and app.go; they just wire Config + configDefaults into the
// generic engine in config.go.
func loadConfig(programName, configName string) (*Config, error) {
	return LoadConfig[Config](getConfigPath(programName), orDefault(configName, "config"), configDefaults)
}

func saveConfig(config *Config, programName, configName string) error {
	return SaveConfig(config, getConfigPath(programName), orDefault(configName, "config"))
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

var programName string = "dirtree"

var (
	jsonTree     bool
	txtTree      bool
	terminalTree bool
	annotateTree bool
	verbose      bool
	saveConf     bool
	displayConf  bool
	filesFirst   bool
	hiddenFiles  bool
	alphabetic   bool
	unrestricted bool
	dirHints     bool
)

var (
	outputPath string
	powerLevel string
	configName string
)

var (
	density            int
	annotationsPadding int
	treeSet            int
	maxDepth           int
	maxElements        int
)

var (
	infoLog    *log.Logger
	debugLog   *log.Logger
	warningLog *log.Logger
	errorLog   *log.Logger
)

func init() {
	// Initialize loggers with different prefixes and flags
	infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	debugLog = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime)
	warningLog = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
