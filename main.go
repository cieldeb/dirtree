package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	// "runtime"
)

var programName string = "dirtree"

var defaultConfig string = `# Dirtree Configuration
jsonTree: false
txtTree: false
terminalTree: true
annotateTree: false
density: 3
annotationsPadding: 3
filesFirst: true
hiddenFiles: false
alphabetic: true
connectorSet: 2
maxDepth: 10
maxElements: 20
dirHints: true
`

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
)

var (
	density            int
	annotationsPadding int
	connectorSet       int
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

	config, err := loadConfig(programName, defaultConfig)
	if err != nil {
		warningLog.Printf("Could not load config: %v", err)
		config = &Config{}
	}

	flag.BoolVar(&jsonTree, "json", config.JsonTree, "Set true to output the tree as JSON")
	flag.BoolVar(&txtTree, "txt", config.TxtTree, "Set true to output the tree in a text file")
	flag.BoolVar(&terminalTree, "terminal", config.TerminalTree, "Set true to output the tree to the terminal")
	flag.BoolVar(&annotateTree, "annotate", config.AnnotateTree, "Set true to add annotations to the tree")
	flag.StringVar(&outputPath, "output", "", "The output directory path")
	flag.IntVar(&density, "density", config.Density, "The tree density, 1 is dense, 5 is spacious")
	flag.IntVar(&annotationsPadding, "annotationspadding", config.AnnotationsPadding, "The padding between annotations fields")
	flag.BoolVar(&filesFirst, "filesfirst", config.FilesFirst, "List files before directories")
	flag.BoolVar(&hiddenFiles, "hiddenfiles", config.HiddenFiles, "Add hidden files to the tree")
	flag.BoolVar(&alphabetic, "alphabetic", config.Alphabetic, "Sort entries alphabetically")
	flag.IntVar(&connectorSet, "connectorset", config.ConnectorSet, "The connector set among \"└├│─\"(1), \"+|-\"(2) and \"|_/\"(3)")
	flag.IntVar(&maxDepth, "maxdepth", config.MaxDepth, "Maximum folder nesting depth the program can go")
	flag.IntVar(&maxElements, "maxelements", config.MaxElements, "Maximum amount of elements shown in a folder")
	flag.BoolVar(&dirHints, "dirhints", config.DirHints, "Add path hints when giving a directory's details")

	flag.StringVar(&outputPath, "output", "", "The output directory path")

	// Config independent parameters
	flag.BoolVar(&verbose, "verbose", false, "Add debug prints to the output")
	flag.BoolVar(&unrestricted, "unrestricted", false, "Bypass restrictions on nesting depth and output length. Things might break !")
	flag.BoolVar(&saveConf, "saveconf", false, "Save current flag values to the configuration file")
	flag.BoolVar(&displayConf, "displayconf", false, "Display the current configuration in the terminal")

	// runtime.GOMAXPROCS(runtime.NumCPU())
}

// TODO Add a flag to enable and disable directory remainders hinting (whether to give the dir path or not)
// TODO Add resources management with runtime, like low power (2 cores), normal (half) and full blast (all cores)
func main() {
	flag.Parse()

	// Get inputPath as a positional argument
	inputPath := flag.Args()[0]
	if inputPath == "" {
		// The user doesn't want to use the program
		if displayConf || saveConf {
			if displayConf {
				displayConfiguration()
			}
			if saveConf {
				saveConfiguration()
			}
			os.Exit(0)
		} else {
			errorLog.Fatal("Please provide a directory to explore as argument")
		}
	}

	// Configure verbose logging output
	if !verbose {
		debugLog.SetOutput(io.Discard)
	}

	// Display config if requested
	if displayConf {
		displayConfiguration()
	}

	// Save config if requested
	if saveConf {
		saveConfiguration()
	}

	// Warning the user when working in unrestricted mode
	if unrestricted {
		infoLog.Println("Unrestricted is set to true. If you continue, the output might be very long and the memory usage heavy. Press ctrl+c to abort or Enter to continue")
		fmt.Scanln()
		maxDepth = 100000
		maxElements = 100000
	}

	tree := NewTree(
		inputPath,
		outputPath,
		terminalTree,
		txtTree,
		jsonTree,
		annotateTree,
		density,
		annotationsPadding,
		connectorSet,
		filesFirst,
		hiddenFiles,
		alphabetic,
		maxDepth,
		maxElements,
		dirHints,
	)
	_ = tree
}

// saveConfiguration saves current configuration
func saveConfiguration() {
	newConfig := &Config{
		JsonTree:           jsonTree,
		TxtTree:            txtTree,
		TerminalTree:       terminalTree,
		AnnotateTree:       annotateTree,
		Density:            density,
		AnnotationsPadding: annotationsPadding,
		FilesFirst:         filesFirst,
		HiddenFiles:        hiddenFiles,
		Alphabetic:         alphabetic,
		ConnectorSet:       connectorSet,
		MaxDepth:           maxDepth,
		MaxElements:        maxElements,
		DirHints:           dirHints,
	}
	if err := saveConfig(newConfig, programName); err != nil {
		errorLog.Fatalf("Error saving config: %s", err)
	}
	infoLog.Println("Configuration saved successfully")
}

// displayConfiguration displays the current configuration in the terminal
func displayConfiguration() {
	infoLog.Printf(`jsonTree %t
		txtTree %t
		terminalTree %t
		annotateTree %t
		density %d
		annotationsPadding %d,
		filesFirst %t,
		hiddenFiles %t,
		alphabetic %t,
		connectorSet %d
		maxDepth %d,
		maxElements %d`, jsonTree, txtTree, terminalTree, annotateTree, density, annotationsPadding, filesFirst, hiddenFiles, alphabetic, connectorSet, maxDepth, maxElements)
}
