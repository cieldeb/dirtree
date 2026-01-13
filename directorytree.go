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

// https://stackoverflow.com/questions/27775376/value-receiver-vs-pointer-receiver

// Tree represents a directory tree generator
type Tree struct {
	inputPath          string
	outputPath         string
	terminalTree       bool
	txtTree            bool
	jsonTree           bool
	annotateTree       bool
	density            int
	annotationsPadding int
	excluded           []string

	// Internal fields
	visualTree bool
	// metadataFields  []string
	imagesMetadata  map[string][]string
	annotated       []string
	jsonStructure   map[string]interface{}
	visualStructure strings.Builder
	maxLineLength   int
	elementIsDir    bool
	depth           int
}

// NewTree creates a new Tree instance
func NewTree(inputPath, outputPath string, terminalTree, txtTree, jsonTree, annotateTree bool, density, annotationsPadding int) *Tree {
	t := &Tree{
		inputPath:          inputPath,
		outputPath:         outputPath,
		terminalTree:       terminalTree,
		txtTree:            txtTree,
		jsonTree:           jsonTree,
		annotateTree:       annotateTree,
		density:            density,
		annotationsPadding: annotationsPadding,
		imagesMetadata:     make(map[string][]string),
		maxLineLength:      0,
		depth:              0,
	}

	if t.outputPath == "" {
		t.outputPath = filepath.Join(t.inputPath, "tree")
	}

	if t.annotateTree {
		t.annotated = []string{"tags"}
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
	}

	t.generate()
	return t
}

// generate builds the tree, annotates it, and writes data to output files
func (t *Tree) generate() {
	if t.jsonTree {
		t.jsonStructure = t.traverseTree(t.inputPath)
	} else {
		t.traverseTree(t.inputPath)
	}

	// Annotate the tree
	//if t.annotateTree {
	//	t.annotate()
	//}

	// Ensure output directory exists
	if err := os.MkdirAll(t.outputPath, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
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

	// Sort entries
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		var baseName string

		if entry.IsDir() {
			baseName = entry.Name()

			// Check if excluded
			if slices.Contains(t.excluded, baseName) {
				fmt.Printf("Skipping excluded : %s\n", baseName)
				if t.jsonTree {
					// Count subdirectories and files
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

			// Process directory
			t.elementIsDir = true
			t.createTreeLine(fullPath)

			if t.visualTree {
				t.depth++
			}

			subtree := t.traverseTree(fullPath)

			if t.jsonTree {
				directoryTree[entry.Name()] = subtree
			}

			if t.visualTree {
				t.depth--
			}
		} else {
			// Process file
			t.elementIsDir = false
			t.createTreeLine(fullPath)

			if t.jsonTree {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
					// Image file - could gather metadata here
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

// annotate adds annotations to the visual tree
//func (t *Tree) annotate() {
//	for i, line := range t.blip {
//		parts := strings.Split(line, " ")
//		fileName := parts[len(parts)-1]
//
//		if tags, exists := t.imagesMetadata[fileName]; exists {
//			tagsText := strings.Join(tags, ", ")
//			padding := t.maxLineLength - len(line)
//			if padding < 0 {
//				padding = 0
//			}
//			spaces := strings.Repeat(" ", padding) + strings.Repeat("  ", t.annotationsPadding)
//			t.blip[i] = line + spaces + tagsText
//		}
//	}
//}

// createTreeLine constructs a line for the tree
func (t *Tree) createTreeLine(path string) {
	dirSuffix := ""
	if t.elementIsDir {
		dirSuffix = "/"
	}

	lineStr := fmt.Sprintf("%s %s%s%s",
		strings.Repeat("|"+strings.Repeat(" ", t.density), t.depth)+"|"+strings.Repeat("_", t.density),
		filepath.Base(path),
		dirSuffix,
		strings.Repeat("  ", t.annotationsPadding),
	)

	if t.visualTree {
		t.visualStructure.WriteString(lineStr + "\n")
		if t.annotateTree {
			lineLen := len(lineStr)
			if lineLen > t.maxLineLength {
				t.maxLineLength = lineLen
			}
		}
	}
}
