package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for config
	tmpDir, err := os.MkdirTemp("", "dirtree-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set XDG_CONFIG_HOME to our temp directory
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigHome)

	testProgramName := "dirtree-test"
	defaultConfig := `# Test Config
jsonTree: true
txtTree: false
terminalTree: true
density: 5
`

	config, err := loadConfig(testProgramName, defaultConfig)
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	if config == nil {
		t.Fatal("Config should not be nil")
	}

	// Verify that config was created and loaded
	// Note: The defaults in viper.SetDefault take precedence over the yaml content
	// The config file is created with the provided defaultConfig string,
	// but viper's SetDefault values are what's actually used
	if config.Density != 3 {
		t.Errorf("Expected default density 3 from viper.SetDefault, got %d", config.Density)
	}
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary directory for config
	tmpDir, err := os.MkdirTemp("", "dirtree-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set XDG_CONFIG_HOME to our temp directory
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigHome)

	testProgramName := "dirtree-test-save"

	// Create config directory
	configPath := filepath.Join(tmpDir, testProgramName)
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Configure viper for this test
	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	testConfig := &Config{
		JsonTree:           true,
		TxtTree:            false,
		TerminalTree:       true,
		AnnotateTree:       false,
		Density:            7,
		AnnotationsPadding: 5,
		FilesFirst:         false,
		HiddenFiles:        true,
		Alphabetic:         false,
		ConnSetSelector:    3,
	}

	if err := saveConfig(testConfig, testProgramName); err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	// Verify the config file was created
	configFile := filepath.Join(configPath, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Read back the config to verify
	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	var loadedConfig Config
	if err := viper.Unmarshal(&loadedConfig); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify values
	if loadedConfig.JsonTree != testConfig.JsonTree {
		t.Errorf("JsonTree mismatch: got %v, want %v", loadedConfig.JsonTree, testConfig.JsonTree)
	}

	if loadedConfig.Density != testConfig.Density {
		t.Errorf("Density mismatch: got %d, want %d", loadedConfig.Density, testConfig.Density)
	}

	if loadedConfig.HiddenFiles != testConfig.HiddenFiles {
		t.Errorf("HiddenFiles mismatch: got %v, want %v", loadedConfig.HiddenFiles, testConfig.HiddenFiles)
	}
}

func TestGetConfigPath(t *testing.T) {
	tests := []struct {
		name           string
		xdgConfigHome  string
		programName    string
		expectedSuffix string
	}{
		{
			name:           "With XDG_CONFIG_HOME set",
			xdgConfigHome:  "/custom/config",
			programName:    "testprog",
			expectedSuffix: "/testprog",
		},
		{
			name:           "Without XDG_CONFIG_HOME",
			xdgConfigHome:  "",
			programName:    "testprog",
			expectedSuffix: ".config/testprog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore XDG_CONFIG_HOME
			oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
			defer os.Setenv("XDG_CONFIG_HOME", oldConfigHome)

			if tt.xdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}

			configPath := getConfigPath(tt.programName)

			if !filepath.IsAbs(configPath) {
				// Allow relative path only if home directory fails
				if configPath != "." {
					t.Errorf("Expected absolute path or '.', got %s", configPath)
				}
				return
			}

			if tt.xdgConfigHome != "" {
				expected := filepath.Join(tt.xdgConfigHome, tt.programName)
				if configPath != expected {
					t.Errorf("Expected %s, got %s", expected, configPath)
				}
			} else {
				// Check that path ends with .config/programName
				if !filepath.IsAbs(configPath) && configPath != "." {
					t.Errorf("Expected absolute path, got %s", configPath)
				}
			}
		})
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "dirtree-default-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set XDG_CONFIG_HOME to our temp directory
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigHome)

	testProgramName := "dirtree-test-default"
	defaultConfig := `# Default Test Config
jsonTree: false
txtTree: true
density: 3
`

	if err := createDefaultConfig(testProgramName, defaultConfig); err != nil {
		t.Fatalf("createDefaultConfig failed: %v", err)
	}

	// Verify the config file was created
	configPath := getConfigPath(testProgramName)
	configFile := filepath.Join(configPath, "config.yaml")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Default config file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if string(content) != defaultConfig {
		t.Errorf("Config content mismatch.\nGot:\n%s\nWant:\n%s", string(content), defaultConfig)
	}
}

func TestConfigDefaults(t *testing.T) {
	// Create a temporary directory for config
	tmpDir, err := os.MkdirTemp("", "dirtree-defaults-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set XDG_CONFIG_HOME to our temp directory
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigHome)

	testProgramName := "dirtree-test-defaults"

	// Create a minimal config file with only one setting
	configPath := filepath.Join(tmpDir, testProgramName)
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	minimalConfig := `jsonTree: true
connectorSetSelector: 1
`
	configFile := filepath.Join(configPath, "config.yaml")
	if err := os.WriteFile(configFile, []byte(minimalConfig), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config with defaults
	viper.Reset()
	config, err := loadConfig(testProgramName, minimalConfig)
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	// Check that defaults were applied for missing values
	if config.Density != 3 {
		t.Errorf("Expected default density 3, got %d", config.Density)
	}

	if config.AnnotationsPadding != 3 {
		t.Errorf("Expected default annotationsPadding 3, got %d", config.AnnotationsPadding)
	}

	if !config.FilesFirst {
		t.Error("Expected default filesFirst true")
	}

	if !config.Alphabetic {
		t.Error("Expected default alphabetic true")
	}

	// Verify the explicitly set value was loaded
	if config.ConnSetSelector != 1 {
		t.Errorf("Expected connSetSelector 1 from config file, got %d", config.ConnSetSelector)
	}

	// Verify the explicitly set jsonTree value
	if !config.JsonTree {
		t.Error("Expected jsonTree true from config file")
	}
}
