package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jphastings/mela-recipes"
)

var (
	version = "0.0.0"
	commit  = "dev"
	date    = time.Now().Format(time.DateOnly)
)

func main() {
	if len(os.Args) < 3 {
		execName := filepath.Base(os.Args[0])
		fmt.Printf(
			"Mela Standardize v%s-%s (%s)\n\nUsage: %s <.melarecipe(s)> [...<.melarecipe(s)>] <output directory>\n",
			version, commit, date, execName)
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
			if err := r.Standardize(true); err != nil {
				fmt.Fprintf(os.Stderr, "Error standardizing '%s' from '%s': %v\n", r.Title, file, err)
				os.Exit(1)
			}

			for _, s := range r.ListStandardizations() {
				fmt.Printf("→ %s\n", s)
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
