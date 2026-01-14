package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// Tree represents a directory tree generator
type Tree struct {
	inputPath       string
	outputPath      string
	terminalTree    bool
	txtTree         bool
	jsonTree        bool
	density         int
	excluded        []string
	connSetSelector int
	filesFirst      bool
	hiddenFiles     bool
	alphabetic      bool
	connectors      connectorSet

	// Internal fields
	visualTree      bool
	jsonStructure   map[string]any
	prefix          string
	visualStructure strings.Builder
	maxLineLength   int
	depth           int
	elementIsDir    bool
}

// Base connector set structure for use in the program
type connectorSet struct {
	branch     string
	leaf       string
	horizontal string
	vertical   string
}

// NewTree creates a new Tree instance
func NewTree(inputPath, outputPath string, terminalTree, txtTree, jsonTree, annotateTree bool, density, annotationsPadding, connSetSelector int, filesFirst bool, hiddenFiles bool, alphabetic bool,
) *Tree {
	t := &Tree{
		inputPath:       inputPath,
		outputPath:      outputPath,
		terminalTree:    terminalTree,
		txtTree:         txtTree,
		jsonTree:        jsonTree,
		density:         density,
		connSetSelector: connSetSelector,
		filesFirst:      filesFirst,
		hiddenFiles:     hiddenFiles,
		alphabetic:      alphabetic,
		maxLineLength:   0,
		depth:           0,
	}

	if t.outputPath == "" {
		t.outputPath = filepath.Join(t.inputPath, "tree")
	}

	// Get excluded folders from user input
	fmt.Print("Excluded subfolders (comma separated, no spaces): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		t.excluded = strings.Split(input, ",")
	}

	if t.jsonTree {
		// Initialize the json map
		t.jsonStructure = make(map[string]any)
	}

	if t.terminalTree || t.txtTree {
		// Initialize the visual structure buffer
		t.visualTree = true
		t.visualStructure.Grow(4096)
		t.prefix = ""

		switch t.connSetSelector {
		case 1:
			t.connectors = connectorSet{
				branch:     "\u251C", // ├ (box drawings light vertical and right)
				leaf:       "\u2514", // └ (box drawings light up and right)
				horizontal: "\u2500", // ─ (box drawings light horizontal)
				vertical:   "\u2502", // │ (box drawings light vertical)
			}
		case 2:
			t.connectors = connectorSet{
				branch:     "+",
				leaf:       "`",
				horizontal: "-",
				vertical:   "|",
			}
		case 3:
			t.connectors = connectorSet{
				branch:     "|",
				leaf:       "|",
				horizontal: "_",
				vertical:   "|",
			}
		default:
			// Default to case 2 if invalid selector
			t.connectors = connectorSet{
				branch:     "+",
				leaf:       "+",
				horizontal: "-",
				vertical:   "|",
			}
		}
	}

	t.generate()
	return t
}

// generate builds the tree writes data to the desired output(s)
func (t *Tree) generate() {
	if t.jsonTree {
		t.jsonStructure = t.traverseTree(t.inputPath)
	} else {
		t.traverseTree(t.inputPath)
	}

	if t.jsonTree || t.txtTree {
		// Ensure output directory exists
		if err := os.MkdirAll(t.outputPath, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			return
		}
	}

	// Write visual tree
	if t.visualTree {
		content := t.visualStructure.String()
		if t.txtTree {
			treeFile := filepath.Join(t.outputPath, "tree.txt")
			if err := os.WriteFile(treeFile, []byte(content), 0644); err != nil {
				fmt.Printf("Error writing visual tree file: %v\n", err)
			}
		}
		if t.terminalTree {
			fmt.Println(content)
		}
	}

	// Write JSON tree
	if t.jsonTree {
		jsonFile := filepath.Join(t.outputPath, "tree.json")
		data, err := json.MarshalIndent(t.jsonStructure, "", "    ")
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
			return
		}
		if err := os.WriteFile(jsonFile, data, 0644); err != nil {
			fmt.Printf("Error writing JSON tree file: %v\n", err)
		}
	}
}

// traverseTree builds the file tree recursively
func (t *Tree) traverseTree(dirPath string) map[string]any {
	directoryTree := make(map[string]any)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", dirPath, err)
		return directoryTree
	}

	// Filter hidden files in place
	n := 0
	for _, entry := range entries {
		if !t.hiddenFiles && entry.Name()[0] == '.' {
			continue
		}
		entries[n] = entry
		n++
	}
	entries = entries[:n]

	// Sort the filtered entries
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() == entries[j].IsDir() {
			if t.alphabetic {
				return entries[i].Name() < entries[j].Name()
			} else {
				return entries[i].Name() > entries[j].Name()
			}
		}
		if t.filesFirst {
			return !entries[i].IsDir() && entries[j].IsDir()
		} else {
			return entries[i].IsDir() && !entries[j].IsDir()
		}
	})

	for idx, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		isLast := idx == len(entries)-1

		if entry.IsDir() {
			baseName := entry.Name()

			// Check if excluded
			if slices.Contains(t.excluded, baseName) {
				fmt.Printf("Skipping excluded: %s\n", baseName)
				if t.jsonTree {
					subEntries, _ := os.ReadDir(fullPath)
					dirs := 0
					files := 0
					for _, subEntry := range subEntries {
						if subEntry.IsDir() {
							dirs++
						} else {
							files++
						}
					}
					dirWord := "y"
					if dirs > 1 {
						dirWord = "ies"
					}
					fileWord := ""
					if files != 1 {
						fileWord = "s"
					}
					directoryTree[entry.Name()] = fmt.Sprintf("%d subdirector%s, %d file%s", dirs, dirWord, files, fileWord)
				}
				continue
			}

			// Check if directory contains only hidden files
			subEntries, err := os.ReadDir(fullPath)
			if err != nil {
				fmt.Printf("Error reading subdirectory %s: %v\n", fullPath, err)
				continue
			}

			// Count non-hidden entries
			nonHiddenCount := 0
			for _, subEntry := range subEntries {
				if t.hiddenFiles || subEntry.Name()[0] != '.' {
					nonHiddenCount++
				}
			}

			// Skip directory if it contains only hidden files and hiddenFiles is false
			if !t.hiddenFiles && nonHiddenCount == 0 {
				continue
			}

			// Process directory
			t.elementIsDir = true
			t.createTreeLine(fullPath, isLast)

			if t.visualTree {
				if isLast {
					t.prefix += "    "
				} else {
					t.prefix += t.connectors.vertical + "   "
				}
			}

			subtree := t.traverseTree(fullPath)

			if t.jsonTree {
				directoryTree[entry.Name()] = subtree
			}

			if t.visualTree {
				if len(t.prefix) >= 4 {
					t.prefix = t.prefix[:len(t.prefix)-4]
				}
			}
		} else {
			if !t.hiddenFiles && entry.Name()[0] == '.' {
				continue
			}
			t.elementIsDir = false
			t.createTreeLine(fullPath, isLast)

			if t.jsonTree {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
					directoryTree[entry.Name()] = map[string]string{
						entry.Name(): entry.Name() + "_data",
					}
				} else {
					directoryTree[entry.Name()] = map[string]any{}
				}
			}
		}
	}

	return directoryTree
}

// createTreeLine constructs a line for the tree
func (t *Tree) createTreeLine(path string, isLast bool) {
	dirSuffix := ""
	if t.elementIsDir {
		dirSuffix = "/"
	}

	// Choose the correct connector based on whether this is the last item
	connector := t.connectors.branch
	if isLast {
		connector = t.connectors.leaf
	}

	lineStr := fmt.Sprintf("%s%s%s %s%s",
		t.prefix,
		connector,
		strings.Repeat(t.connectors.horizontal, t.density),
		filepath.Base(path),
		dirSuffix,
	)

	if t.visualTree {
		t.visualStructure.WriteString(lineStr + "\n")
	}
}
