// SPDX-License-Identifier: GPL-3.0-or-later
// TODO maybe split this file :sweat:
package main

import (
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
	inputPath    string
	outputPath   string
	terminalTree bool
	txtTree      bool
	jsonTree     bool
	density      int
	excluded     []string
	connectorSet int
	filesFirst   bool
	hiddenFiles  bool
	alphabetic   bool
	connectors   connectors

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
type connectors struct {
	branch     string
	leaf       string
	horizontal string
	vertical   string
}

// NewTree creates a new Tree instance
func NewTree(inputPath, outputPath string, config *Config) *Tree {
	t := &Tree{
		inputPath:     inputPath,
		outputPath:    outputPath,
		terminalTree:  config.TerminalTree,
		txtTree:       config.TxtTree,
		jsonTree:      config.JsonTree,
		density:       config.Density,
		connectorSet:  config.TreeSet,
		filesFirst:    config.FilesFirst,
		hiddenFiles:   config.HiddenFiles,
		alphabetic:    config.Alphabetic,
		maxLineLength: 0,
		depth:         0,
	}

	if t.outputPath == "" {
		t.outputPath = filepath.Join(t.inputPath)
	}

	// Get excluded folders from user input
	// fmt.Print("Excluded subfolders (comma separated, no spaces): ")
	// reader := bufio.NewReader(os.Stdin)
	// input, _ := reader.ReadString('\n')
	// input = strings.TrimSpace(input)
	// if input != "" {
	// 	t.excluded = strings.Split(input, ",")
	// }

	if t.jsonTree {
		// Initialize the json map
		t.jsonStructure = make(map[string]any)
	}

	if t.terminalTree || t.txtTree {
		// Initialize the visual structure buffer
		t.visualTree = true
		t.visualStructure.Grow(4096)
		t.prefix = ""

		switch t.connectorSet {
		case 1:
			t.connectors = connectors{
				branch:     "├",
				leaf:       "└",
				horizontal: "─",
				vertical:   "│",
			}
		case 2:
			t.connectors = connectors{
				branch:     "+",
				leaf:       "`",
				horizontal: "-",
				vertical:   "|",
			}
		case 3:
			t.connectors = connectors{
				branch:     "|",
				leaf:       "|",
				horizontal: "_",
				vertical:   "|",
			}
		case 4:
			t.connectors = connectors{
				branch:     "o",
				leaf:       "o.",
				horizontal: ".",
				vertical:   "o",
			}
		case 5:
			t.connectors = connectors{
				branch:     "║",
				leaf:       "╚",
				horizontal: "═",
				vertical:   "║",
			}
		case 6:
			t.connectors = connectors{
				branch:     "╿",
				leaf:       "┖",
				horizontal: "╼",
				vertical:   "╿",
			}
		case 7:
			t.connectors = connectors{
				branch:     "│",
				leaf:       "╰",
				horizontal: "─",
				vertical:   "│",
			}
		default:
			// Default to case 2 if invalid selector
			t.connectors = connectors{
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

// generate builds the tree and writes data to the desired output(s)
func (t *Tree) generate() {
	var emptyDir []os.DirEntry
	if t.jsonTree {
		t.jsonStructure = t.traverseTree(t.inputPath, &emptyDir, maxDepth, maxElements)
	} else {
		t.traverseTree(t.inputPath, &emptyDir, maxDepth, maxElements)
	}

	// Write visual tree
	if t.visualTree {
		content := t.visualStructure.String()
		if t.txtTree {
			treeFile := filepath.Join(t.outputPath, "tree.txt")
			contentWithBOM := append([]byte{0xEF, 0xBB, 0xBF}, []byte(t.inputPath+"\n")...)
			contentWithBOM = append(contentWithBOM, []byte(content)...)
			if err := os.WriteFile(treeFile, contentWithBOM, 0644); err != nil {
				fmt.Printf("Error writing visual tree file: %v\n", err)
			}
		}
		if t.terminalTree {
			os.Stdout.Write([]byte(content + "\n"))
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
func (t *Tree) traverseTree(dirPath string, entries *[]os.DirEntry, maxDepth, maxElements int) map[string]any {
	directoryTree := make(map[string]any)

	localEntries := *entries

	if len(localEntries) == 0 {
		readEntries, err := os.ReadDir(dirPath)
		if err != nil {
			fmt.Printf("Error reading directory %s: %v\n", dirPath, err)
			return directoryTree
		}
		localEntries = readEntries
	}

	// Filter hidden files in place
	n := 0
	for _, entry := range localEntries {
		if !t.hiddenFiles && entry.Name()[0] == '.' {
			continue
		}
		localEntries[n] = entry
		n++
	}
	localEntries = localEntries[:n]

	// Sort the filtered entries
	sort.Slice(localEntries, func(i, j int) bool {
		if localEntries[i].IsDir() == localEntries[j].IsDir() {
			if t.alphabetic {
				return localEntries[i].Name() < localEntries[j].Name()
			} else {
				return localEntries[i].Name() > localEntries[j].Name()
			}
		}
		if t.filesFirst {
			return !localEntries[i].IsDir() && localEntries[j].IsDir()
		} else {
			return localEntries[i].IsDir() && !localEntries[j].IsDir()
		}
	})

	for idx, entry := range localEntries {
		fullPath := filepath.Join(dirPath, entry.Name())
		isLast := idx == len(localEntries)-1

		// If we reach the maximum amount of elements, we create an additional line describing what is left in the directory
		if idx == maxElements {
			if t.visualTree {
				t.visualStructure.WriteString(t.composeSummary(dirPath, localEntries, idx, false, isLast) + "\n")
			}
			break
		}

		if entry.IsDir() {
			baseName := entry.Name()

			// Reading the entries contained by the folder
			subEntries, err := os.ReadDir(fullPath)
			if err != nil {
				fmt.Printf("Error reading subdirectory %s: %v\n", fullPath, err)
				continue
			}

			// Check if excluded
			if slices.Contains(t.excluded, baseName) {
				fmt.Printf("Skipping excluded: %s\n", baseName)
				if t.jsonTree {
					// Default case with 0 start index of entries reading
					directoryTree[entry.Name()] = t.composeSummary(dirPath, subEntries, 0, true, isLast)
					continue
				}
			}

			// Count specific elements amounts if relevant
			var nonHiddenCount int
			if !t.hiddenFiles {
				for _, subEntry := range subEntries {
					// Count non-hidden entries
					if t.hiddenFiles || subEntry.Name()[0] != '.' {
						nonHiddenCount++
					}
				}
			}

			// Skip directory if it contains only hidden files and hiddenFiles is false
			if !t.hiddenFiles && nonHiddenCount == 0 {
				continue
			}

			// Process directory
			t.elementIsDir = true
			t.createTreeLine(fullPath, isLast)

			// Checking that we didn't exceed the maximum nesting depth
			if t.depth+1 <= maxDepth {
				t.depth += 1
				if t.visualTree {
					// Composing the prefix for the directory
					prefixAddition := ""
					if isLast {
						prefixAddition = strings.Repeat(" ", t.density)
					} else {
						prefixAddition = t.connectors.vertical + strings.Repeat(" ", t.density)
					}
					t.prefix += prefixAddition

					subtree := t.traverseTree(fullPath, &subEntries, maxDepth, maxElements)
					if t.jsonTree {
						directoryTree[entry.Name()] = subtree
					}

					// Trimming the prefix after having processed the directory
					t.prefix = t.prefix[:len(t.prefix)-len(prefixAddition)]
				} else {
					subtree := t.traverseTree(fullPath, &subEntries, maxDepth, maxElements)
					if t.jsonTree {
						directoryTree[entry.Name()] = subtree
					}
				}
				t.depth -= 1
			} else {
				if t.visualTree {
					t.visualStructure.WriteString(t.composeSummary(fullPath, subEntries, 0, false, isLast) + "\n")
				}
			}
		} else {
			// Skipping hidden entries
			if !t.hiddenFiles && entry.Name()[0] == '.' {
				continue
			}

			t.elementIsDir = false
			t.createTreeLine(fullPath, isLast)

			// Handling the case of json tree
			if t.jsonTree {
				directoryTree[entry.Name()] = map[string]any{}
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

// composeSummary creates the summary of a folder that exceeds the limits set by the user.
// If maxElements is reached for a directory, a final string is placed as the last line for the folder,
// string that can look like that : "and 5 subdirectories and 1 file not shown, full directory path : /path/to/directory"
func (t *Tree) composeSummary(dirPath string, entries []os.DirEntry, startidx int, jsonCase, isLast bool) string {
	dirs := 0
	files := 0

	// Counting remaining files and directories
	for i := startidx; i < len(entries); i++ {
		if entries[i].IsDir() {
			dirs++
		} else {
			files++
		}
	}

	// Creating and building the string base
	var b strings.Builder

	// Add the prefix for the current depth
	b.WriteString(t.prefix)

	// Adding spaces if we are at maximum nesting depth and not maximum elements amount
	if t.depth == maxDepth && startidx == 0 {
		// If the element is not the last line describing the directory, adding the vertical connector
		if !isLast {
			b.WriteString(t.connectors.vertical)
		} else {
			b.WriteString(" ")
		}
		// In any case, adding spaces to show nesting of the directory
		b.WriteString(strings.Repeat(" ", t.density))
	}

	// Adding the base of the branch, the connectors and a space
	b.WriteString(t.connectors.leaf)
	b.WriteString(strings.Repeat(t.connectors.horizontal, t.density))
	b.WriteString(" ")

	// Set the order in which elements are added to the description string depending on the filesFirst flag value
	var firstElementName, secondElementName string
	var firstQuantity, secondQuantity int

	if filesFirst {
		firstElementName = "file"
		secondElementName = "subdirector"
		firstQuantity = files
		secondQuantity = dirs
	} else {
		firstElementName = "subdirector"
		secondElementName = "file"
		firstQuantity = dirs
		secondQuantity = files
	}

	// Add plurals
	firstElementName = pluralize(firstElementName, firstQuantity)
	secondElementName = pluralize(secondElementName, secondQuantity)

	// Create string parts for subdirectories count
	if firstQuantity != 0 {
		fmt.Fprintf(&b, "%d %s", firstQuantity, firstElementName)
	}

	// If needed, add connecting "and"
	if firstQuantity != 0 && secondQuantity != 0 {
		b.WriteString(" and ")
	}

	// Create string parts for files count
	if secondQuantity != 0 {
		fmt.Fprintf(&b, "%d %s", secondQuantity, secondElementName)
	}

	// Add sentence end when we are not purely descriptive
	if !jsonCase {
		fmt.Fprint(&b, " not shown")
		if t.depth != 0 && dirHints {
			fmt.Fprintf(&b, ", full directory path : %s", dirPath)
		}
	}

	return b.String()
}
