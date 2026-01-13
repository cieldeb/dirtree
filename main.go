package main

import (
	"flag"
	"log"
)

var (
	jsonTree     bool
	txtTree      bool
	terminalTree bool
	annotateTree bool
)

var (
	inputPath  string
	outputPath string
)

var (
	density            int
	annotationsPadding int
)

func init() {
	flag.BoolVar(&jsonTree, "json", false, "Set true to output the tree as JSON")
	flag.BoolVar(&txtTree, "txt", true, "Set true to output the tree in a text file")
	flag.BoolVar(&terminalTree, "terminal", false, "Set true to output the tree to the terminal")
	flag.BoolVar(&annotateTree, "annotate", false, "Set true to add annotations to the tree")
	flag.StringVar(&inputPath, "input", ".", "The input directory path")
	flag.StringVar(&outputPath, "output", "", "The output directory path")
	flag.IntVar(&density, "density", 3, "The tree density, 1 is dense, 5 is spacious")
	flag.IntVar(&annotationsPadding, "annotationspadding", 3, "The padding between annotations fields")
}

func main() {
	flag.Parse()
	if inputPath == "" {
		log.Fatal("Please provide a directory to explore as argument")
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
	)
	_ = tree
}
