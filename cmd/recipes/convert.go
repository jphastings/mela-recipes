package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jphastings/recipes"
	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/standardize"
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert recipes between different formats",
	Long:  longExplain(),
	RunE: func(cmd *cobra.Command, args []string) error {
		o, err := formats.LoadOptions()
		if err != nil {
			return err
		}

		rs, rc, err := recipes.ParseAll(args, o)
		if err != nil {
			return err
		}

		to, err := cmd.Flags().GetString("to")
		if err != nil {
			return err
		}
		filename, asType, destFormat := recipes.ParseDestination(to)

		makeCollection := asType == recipes.AsTypeCollection || asType == recipes.AsTypeAny && rc != nil
		if makeCollection {
			rcName := "Recipes"
			if rc != nil {
				rcName = rc.Name()
			}

			out, err := destFormat.NewCollection(rcName, rc.Recipes())
			if err != nil {
				return fmt.Errorf("unable to create a new collection in the %s format: %w", destFormat.Name, err)
			}

			if filename == "" {
				filename = standardize.StringToFilename(rcName) + destFormat.ExtensionCollection
			}

			f, err := os.Create(filename)
			if err != nil {
				return fmt.Errorf("unable to make output recipe collection: %w", err)
			}

			return out.Marshal(f)
		}

		fmt.Println(len(rs), rc)
		return fmt.Errorf("convert command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().String("to", "", "The filename, extension, or format to convert into")
	convertCmd.MarkFlagRequired("to")
}

func longExplain() string {
	str := "Converts cookbook recipes between supported formats. Supported formats include:\n"

	for _, f := range recipes.AvailableFormats() {
		str += "\n"
		str += f.Name + "\n"
		str += "  extensions: " + strings.Join(f.Extensions(), ", ") + "\n"
		if f.Features.ParseRecipe || f.Features.ParseCollection {
			str += "  can import?  ✅\n"
		} else {
			str += "  can import?  ❌\n"
		}
		if f.Features.WriteRecipe {
			str += "  can export?  ✅\n"
		} else {
			str += "  can export?  ❌\n"
		}
		if f.Features.WriteCollection {
			str += "  collections? ✅\n"
		} else {
			str += "  collections? ❌\n"
		}
		str += "  more info: " + f.URL + "\n"
	}

	return str
}
