package main

import (
	"fmt"
	"os"

	_ "github.com/jphastings/recipes/cooklang"
	_ "github.com/jphastings/recipes/crouton"
	_ "github.com/jphastings/recipes/epub"
	_ "github.com/jphastings/recipes/mela"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "recipes",
	Short: "Cookbook recipe format toolbox",
	Long:  `Tools for parsing, building, and converting cookbook recipe formats. Supports Mela, Crouton, Paprika, and Cooklang formats as well as extracting from ePub cook books.`,
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "The command failed: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().Bool("out-here", false, "Output files in the current working directory")
	rootCmd.Flags().Bool("out-there", true, "Output files in the same directory as the source data")
	rootCmd.Flags().String("out-dir", "", "Output files to the given directory")
	rootCmd.MarkFlagsMutuallyExclusive("out-here", "out-there", "out-dir")
}
