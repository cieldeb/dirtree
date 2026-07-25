// SPDX-License-Identifier: GPL-3.0-or-later
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	flag "github.com/spf13/pflag"
)

// TODO add a summary mode that has just enough but not too much details and intelligently documents folders
// TODO put excluded elements in a flag
// TODO switch to cobra for config and flags management

func main() {
	// First parsing of flags in case the configuration to be loaded isnt the default one
	preParse := flag.NewFlagSet("preparse", flag.ContinueOnError)
	preParse.ParseErrorsAllowlist.UnknownFlags = true
	preParse.StringVarP(&configName, "configuration", "c", "", "Configuration name if not using the default")
	preParse.Parse(os.Args[1:])

	// Check whether this preset's config file already exists on disk *before*
	// loading it — loadConfig returns configDefaults-backed values either way,
	// so this is the only way to tell "existing preset" from "first run".
	configFilePath := filepath.Join(getConfigPath(programName), orDefault(configName, "config")+".yaml")
	configExisted := true
	if _, statErr := os.Stat(configFilePath); os.IsNotExist(statErr) {
		configExisted = false
	}

	// Loading configuration. loadedCfg is local: it only exists to seed flag
	// defaults below, nothing outside main() needs it.
	loadedCfg, err := loadConfig(programName, configName)
	if err != nil {
		warningLog.Printf("Could not load config: %v", err)
		loadedCfg = &Config{}
	}

	flag.StringVarP(&configName, "configuration", "c", "", "Configuration name if not using the default")
	flag.BoolVar(&jsonTree, "json", loadedCfg.JsonTree, "Set true to output the tree as JSON")
	flag.BoolVar(&txtTree, "txt", loadedCfg.TxtTree, "Set true to output the tree in a text file")
	flag.BoolVar(&terminalTree, "terminal", loadedCfg.TerminalTree, "Set true to output the tree to the terminal")
	flag.BoolVarP(&annotateTree, "annotate", "a", loadedCfg.AnnotateTree, "Set true to add annotations to the tree")
	flag.IntVar(&density, "density", loadedCfg.Density, "The tree density, 1 is dense, 5 is spacious")
	flag.IntVar(&annotationsPadding, "annotationspadding", loadedCfg.AnnotationsPadding, "The padding between annotations fields")
	flag.BoolVar(&filesFirst, "filesfirst", loadedCfg.FilesFirst, "List files before directories")
	flag.BoolVarP(&hiddenFiles, "hidden", "h", loadedCfg.HiddenFiles, "Add hidden files to the tree")
	flag.BoolVar(&alphabetic, "alphabetic", loadedCfg.Alphabetic, "Sort entries alphabetically")
	flag.IntVarP(&treeSet, "treeset", "t", loadedCfg.TreeSet, "The tree connectors set among:\n(1) └ ├ │ ─\n(2) + ` | - \n(3) | | _ |\n(4) o o. . o\n(5) ║ ╚ ═ ║\n(6) ╿ ┖ ╼ ╿\n(7) │ ╰ ─ │\n")
	flag.IntVar(&maxDepth, "maxdepth", loadedCfg.MaxDepth, "Maximum folder nesting depth the program can go")
	flag.IntVar(&maxElements, "maxelements", loadedCfg.MaxElements, "Maximum amount of elements shown in a folder")
	flag.BoolVar(&dirHints, "dirhints", loadedCfg.DirHints, "Add path hints when giving a directory's details")
	flag.StringVar(&powerLevel, "powerLevel", loadedCfg.PowerLevel, "Sets the amount of cpu cores used : l(ow) -> 2, m(edium) -> half of all, a(ll) -> all available cores")

	flag.StringVarP(&outputPath, "output", "o", "", "The output directory path")

	// Config independent parameters
	flag.BoolVarP(&verbose, "verbose", "v", false, "Add debug prints to the output")
	flag.BoolVarP(&unrestricted, "unrestricted", "u", false, "Bypass restrictions on nesting depth, output length and uses all available cpu cores. Things might break !")
	flag.BoolVarP(&saveConf, "saveconf", "s", false, "Save current flag values to the configuration file")
	flag.BoolVar(&displayConf, "displayconf", false, "Display the current configuration in the terminal")

	flag.Parse()

	// Setting number of cpu cores to be used
	switch powerLevel {
	case "l":
		runtime.GOMAXPROCS(2)
	case "m":
		runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	case "a":
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	cfg := &Config{
		JsonTree:           jsonTree,
		TxtTree:            txtTree,
		TerminalTree:       terminalTree,
		AnnotateTree:       annotateTree,
		Density:            density,
		AnnotationsPadding: annotationsPadding,
		FilesFirst:         filesFirst,
		HiddenFiles:        hiddenFiles,
		Alphabetic:         alphabetic,
		TreeSet:            treeSet,
		MaxDepth:           maxDepth,
		MaxElements:        maxElements,
		DirHints:           dirHints,
		PowerLevel:         powerLevel,
	}

	// Display config if requested
	if displayConf {
		infoLog.Printf("%+v", cfg)
	}

	// Save config if explicitly requested, or bootstrap it from the values
	// just parsed if this preset didn't exist yet. When no flag overrides
	// anything (bare `--saveconfig`), newConfig already equals the defaults,
	// so the bootstrap and the explicit-save path naturally produce the same
	// "base config" result without needing to special-case it.
	if saveConf || !configExisted {
		if err := saveConfig(cfg, programName, configName); err != nil {
			errorLog.Fatalf("Error saving config: %s", err)
		}
		if saveConf {
			infoLog.Println("Configuration saved successfully")
		} else {
			infoLog.Printf("Created new configuration %q from provided values", orDefault(configName, "config"))
		}
	}

	// Setting number of cpu cores to be used
	switch powerLevel {
	case "l":
		runtime.GOMAXPROCS(2)
	case "m":
		runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	case "a":
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// Get inputPath as a positional argument
	var inputPath string
	args := flag.Args()
	if len(args) != 0 {
		inputPath = args[0]
	} else {
		// If no path was provided when calling the executable, we default to the current working directory
		cwd, err := os.Getwd()
		if err != nil {
			errorLog.Fatal("Error fetching the current working directory, please provide a path to document.")
		}
		inputPath = cwd
	}

	// Configure verbose logging output
	if !verbose {
		debugLog.SetOutput(io.Discard)
	}

	tree := NewTree(inputPath, outputPath, cfg)
	_ = tree
}
