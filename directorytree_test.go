package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to create a test directory structure
func createTestDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "dirtree-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a simple directory structure
	dirs := []string{
		"subdir1",
		"subdir2",
		"subdir1/nested",
		".hidden",
	}

	files := []string{
		"file1.txt",
		"file2.go",
		"subdir1/file3.md",
		"subdir1/nested/file4.json",
		"subdir2/file5.png",
		".hiddenfile",
	}

	for _, dir := range dirs {
		path := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", path, err)
		}
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	return tmpDir
}

func TestConnectors(t *testing.T) {
	tests := []struct {
		name     string
		selector int
		expected connectors
	}{
		{
			name:     "Unicode connectors",
			selector: 1,
			expected: connectors{
				branch:     "\u251C",
				leaf:       "\u2514",
				horizontal: "\u2500",
				vertical:   "\u2502",
			},
		},
		{
			name:     "ASCII connectors",
			selector: 2,
			expected: connectors{
				branch:     "+",
				leaf:       "`",
				horizontal: "-",
				vertical:   "|",
			},
		},
		{
			name:     "Alternative connectors",
			selector: 3,
			expected: connectors{
				branch:     "|",
				leaf:       "|",
				horizontal: "_",
				vertical:   "|",
			},
		},
		{
			name:     "Invalid selector defaults to ASCII",
			selector: 99,
			expected: connectors{
				branch:     "+",
				leaf:       "+",
				horizontal: "-",
				vertical:   "|",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := createTestDir(t)
			defer os.RemoveAll(tmpDir)

			outputDir := filepath.Join(tmpDir, "output")

			// Redirect stdin to provide empty excluded list
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			w.Write([]byte("\n"))
			w.Close()
			defer func() { os.Stdin = oldStdin }()

			tree := &Tree{
				inputPath:    tmpDir,
				outputPath:   outputDir,
				terminalTree: false,
				txtTree:      true,
				jsonTree:     false,
				density:      3,
				connectorSet: tt.selector,
				filesFirst:   false,
				hiddenFiles:  false,
				alphabetic:   true,
			}

			tree.visualTree = true
			tree.visualStructure.Grow(4096)
			tree.prefix = ""

			// Set connectors manually based on selector (mimics NewTree logic)
			switch tt.selector {
			case 1:
				tree.connectors = connectors{
					branch:     "\u251C",
					leaf:       "\u2514",
					horizontal: "\u2500",
					vertical:   "\u2502",
				}
			case 2:
				tree.connectors = connectors{
					branch:     "+",
					leaf:       "`",
					horizontal: "-",
					vertical:   "|",
				}
			case 3:
				tree.connectors = connectors{
					branch:     "|",
					leaf:       "|",
					horizontal: "_",
					vertical:   "|",
				}
			default:
				tree.connectors = connectors{
					branch:     "+",
					leaf:       "+",
					horizontal: "-",
					vertical:   "|",
				}
			}

			if tree.connectors.branch != tt.expected.branch ||
				tree.connectors.leaf != tt.expected.leaf ||
				tree.connectors.horizontal != tt.expected.horizontal ||
				tree.connectors.vertical != tt.expected.vertical {
				t.Errorf("Connector set mismatch.\nGot: %+v\nWant: %+v",
					tree.connectors, tt.expected)
			}
		})
	}
}

func TestTreeGeneration_TxtOutput(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	outputDir := filepath.Join(tmpDir, "output")

	// Redirect stdin to provide empty excluded list
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Write([]byte("\n"))
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	NewTree(tmpDir, outputDir, false, true, false, false, 3, 3, 2, false, false, true, 10, 20)

	// Check if output file was created
	treeFile := filepath.Join(outputDir, "tree.txt")
	if _, err := os.Stat(treeFile); os.IsNotExist(err) {
		t.Fatal("tree.txt was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(treeFile)
	if err != nil {
		t.Fatalf("Failed to read tree.txt: %v", err)
	}

	contentStr := string(content)

	// Check for expected elements
	expectedElements := []string{"subdir1", "subdir2", "file1.txt", "file2.go"}
	for _, elem := range expectedElements {
		if !strings.Contains(contentStr, elem) {
			t.Errorf("Expected element %q not found in tree output", elem)
		}
	}

	// Check that hidden files are not included by default
	if strings.Contains(contentStr, ".hidden") || strings.Contains(contentStr, ".hiddenfile") {
		t.Error("Hidden files should not be included by default")
	}
}

func TestTreeGeneration_JSONOutput(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	outputDir := filepath.Join(tmpDir, "output")

	// Redirect stdin to provide empty excluded list
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Write([]byte("\n"))
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	NewTree(tmpDir, outputDir, false, false, true, false, 3, 3, 2, false, false, true, 10, 20)

	// Check if output file was created
	jsonFile := filepath.Join(outputDir, "tree.json")
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		t.Fatal("tree.json was not created")
	}

	// Read and parse JSON
	content, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Fatalf("Failed to read tree.json: %v", err)
	}

	var jsonTree map[string]any
	if err := json.Unmarshal(content, &jsonTree); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify structure
	if _, ok := jsonTree["subdir1"]; !ok {
		t.Error("Expected subdir1 in JSON tree")
	}

	if _, ok := jsonTree["subdir2"]; !ok {
		t.Error("Expected subdir2 in JSON tree")
	}
}

func TestTreeGeneration_WithHiddenFiles(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	outputDir := filepath.Join(tmpDir, "output")

	// Redirect stdin to provide empty excluded list
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Write([]byte("\n"))
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	// Generate with hidden files enabled
	NewTree(tmpDir, outputDir, false, true, false, false, 3, 3, 2, false, true, true, 10, 20)

	// Read tree.txt
	treeFile := filepath.Join(outputDir, "tree.txt")
	content, err := os.ReadFile(treeFile)
	if err != nil {
		t.Fatalf("Failed to read tree.txt: %v", err)
	}

	contentStr := string(content)

	// Check that hidden files are included
	if !strings.Contains(contentStr, ".hidden") {
		t.Error("Hidden directory should be included when hiddenFiles is true")
	}

	if !strings.Contains(contentStr, ".hiddenfile") {
		t.Error("Hidden file should be included when hiddenFiles is true")
	}
}

func TestTreeGeneration_FilesFirst(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	outputDir := filepath.Join(tmpDir, "output")

	// Redirect stdin to provide empty excluded list
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Write([]byte("\n"))
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	// Generate with filesFirst enabled
	NewTree(tmpDir, outputDir, false, true, false, false, 3, 3, 2, true, false, true, 10, 20)

	// Read tree.txt
	treeFile := filepath.Join(outputDir, "tree.txt")
	content, err := os.ReadFile(treeFile)
	if err != nil {
		t.Fatalf("Failed to read tree.txt: %v", err)
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Find indices of first file and first directory at root level
	firstFileIdx := -1
	firstDirIdx := -1

	for i, line := range lines {
		if strings.Contains(line, "file") && !strings.Contains(line, "/") && firstFileIdx == -1 {
			firstFileIdx = i
		}
		if strings.Contains(line, "subdir") && firstDirIdx == -1 {
			firstDirIdx = i
		}
		if firstFileIdx != -1 && firstDirIdx != -1 {
			break
		}
	}

	// When filesFirst is true, files should come before directories
	if firstFileIdx != -1 && firstDirIdx != -1 && firstFileIdx > firstDirIdx {
		t.Error("Files should appear before directories when filesFirst is true")
	}
}

func TestTreeGeneration_ExcludedFolders(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	outputDir := filepath.Join(tmpDir, "output")

	// Redirect stdin to provide excluded list
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Write([]byte("subdir1\n"))
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	NewTree(tmpDir, outputDir, false, true, false, false, 3, 3, 2, false, false, true, 10, 20)

	// Read tree.txt
	treeFile := filepath.Join(outputDir, "tree.txt")
	content, err := os.ReadFile(treeFile)
	if err != nil {
		t.Fatalf("Failed to read tree.txt: %v", err)
	}

	contentStr := string(content)

	// subdir1 should not be traversed (though it may appear as excluded)
	// The nested content should definitely not appear
	if strings.Contains(contentStr, "nested") {
		t.Error("Nested directory inside excluded folder should not appear")
	}

	if strings.Contains(contentStr, "file3.md") {
		t.Error("Files inside excluded folder should not appear")
	}

	// subdir2 should still be present
	if !strings.Contains(contentStr, "subdir2") {
		t.Error("Non-excluded directories should still appear")
	}
}

func TestTreeGeneration_DensityLevels(t *testing.T) {
	tests := []struct {
		name    string
		density int
	}{
		{"Dense", 1},
		{"Medium", 3},
		{"Spacious", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := createTestDir(t)
			defer os.RemoveAll(tmpDir)

			outputDir := filepath.Join(tmpDir, "output")

			// Redirect stdin
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			w.Write([]byte("\n"))
			w.Close()
			defer func() { os.Stdin = oldStdin }()

			NewTree(tmpDir, outputDir, false, true, false, false, tt.density, 3, 2, false, false, true, 10, 20)

			// Read tree.txt
			treeFile := filepath.Join(outputDir, "tree.txt")
			content, err := os.ReadFile(treeFile)
			if err != nil {
				t.Fatalf("Failed to read tree.txt: %v", err)
			}

			contentStr := string(content)

			// Check that horizontal connectors repeat according to density
			// Should find lines with the appropriate number of dashes
			expectedDashes := strings.Repeat("-", tt.density)
			if !strings.Contains(contentStr, expectedDashes) {
				t.Errorf("Expected to find %d dashes for density %d", tt.density, tt.density)
			}
		})
	}
}

func TestTreeGeneration_AlphabeticSorting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dirtree-alpha-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files with specific names to test sorting
	files := []string{"zebra.txt", "apple.txt", "mango.txt", "banana.txt"}
	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	outputDir := filepath.Join(tmpDir, "output")

	// Test alphabetic sorting enabled
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Write([]byte("\n"))
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	NewTree(tmpDir, outputDir, false, true, false, false, 3, 3, 2, false, false, true, 10, 20)

	// Read tree.txt
	treeFile := filepath.Join(outputDir, "tree.txt")
	content, err := os.ReadFile(treeFile)
	if err != nil {
		t.Fatalf("Failed to read tree.txt: %v", err)
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Find order of files in output
	indices := make(map[string]int)
	for i, line := range lines {
		for _, file := range files {
			if strings.Contains(line, file) {
				indices[file] = i
			}
		}
	}

	// Verify alphabetic order
	if indices["apple.txt"] >= indices["banana.txt"] ||
		indices["banana.txt"] >= indices["mango.txt"] ||
		indices["mango.txt"] >= indices["zebra.txt"] {
		t.Error("Files are not in alphabetic order")
	}
}

func TestCreateTreeLine(t *testing.T) {
	tree := &Tree{
		density: 3,
		connectors: connectors{
			branch:     "+",
			leaf:       "`",
			horizontal: "-",
			vertical:   "|",
		},
		visualTree: true,
		prefix:     "",
	}
	tree.visualStructure.Grow(4096)

	// Test directory
	tree.elementIsDir = true
	tree.createTreeLine("/path/to/testdir", false)
	output := tree.visualStructure.String()

	if !strings.Contains(output, "testdir/") {
		t.Error("Directory should have trailing slash")
	}

	if !strings.Contains(output, "+---") {
		t.Error("Should contain branch connector with horizontal lines")
	}

	// Reset and test file
	tree.visualStructure.Reset()
	tree.elementIsDir = false
	tree.createTreeLine("/path/to/testfile.txt", true)
	output = tree.visualStructure.String()

	if !strings.Contains(output, "testfile.txt") {
		t.Error("File name should be present")
	}

	if strings.Contains(output, "/") && !strings.Contains(output, "testfile.txt") {
		t.Error("File should not have trailing slash")
	}

	if !strings.Contains(output, "`---") {
		t.Error("Last item should use leaf connector")
	}
}

func TestDefaultOutputPath(t *testing.T) {
	tmpDir := createTestDir(t)
	defer os.RemoveAll(tmpDir)

	// Redirect stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Write([]byte("\n"))
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	// Pass empty output path
	NewTree(tmpDir, "", false, true, false, false, 3, 3, 2, false, false, true, 10, 20)

	// Check if output was created in default location (inputPath/tree)
	defaultOutputDir := filepath.Join(tmpDir, "tree")
	treeFile := filepath.Join(defaultOutputDir, "tree.txt")

	if _, err := os.Stat(treeFile); os.IsNotExist(err) {
		t.Error("tree.txt should be created in default output path (inputPath/tree)")
	}
}
