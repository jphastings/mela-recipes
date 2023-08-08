package main

import (
	"fmt"
	"os"

	"github.com/jphastings/mela-recipes"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <.melarecipe(s) file> [... <.melarecipe(s) file>] <output directory>\n", os.Args[0])
		os.Exit(1)
	}

	inputFiles := os.Args[1 : len(os.Args)-1]
	outputDir := os.Args[len(os.Args)-1]

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Output directory '%s' does not exist\n", outputDir)
		os.Exit(1)
	}

	for _, file := range inputFiles {
		recipes, err := mela.Open(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening '%s': %v\n", file, err)
			os.Exit(1)
		}

		for _, r := range recipes {
			if err := r.Standardize(); err != nil {
				fmt.Fprintf(os.Stderr, "Error standardizing '%s' from '%s': %v\n", r.Title, file, err)
				os.Exit(1)
			}

			dest, err := r.Save(outputDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error saving '%s' from '%s': %v\n", r.Title, file, err)
				os.Exit(1)
			}

			fmt.Printf("Saved '%s' to '%s'\n", r.Title, dest)
		}
	}
}
