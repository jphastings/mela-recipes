package main

import (
	"fmt"
	"strings"

	"github.com/jphastings/recipes"
	"github.com/jphastings/recipes/internal/formats"
	"github.com/spf13/cobra"

	"github.com/schollz/progressbar/v3"
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

		to, err := cmd.Flags().GetString("to")
		if err != nil {
			return err
		}
		filename, asType, destFormat := recipes.ParseDestination(to)

		pe, cd, err := recipes.ParseAll(args, o)
		if err != nil {
			return err
		}

		collectionRequested := asType == recipes.AsTypeCollection || asType == recipes.AsTypeAny && cd != nil
		if collectionRequested {
			if cd == nil {
				cd = &formats.CollectionDetails{}
			}
			if filename != "" {
				cd.Filename = filename
			}
			return makeCollection(cd, destFormat, pe)
		} else {
			return makeRecipes(filename, destFormat, pe)
		}
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

func makeCollection(cd *formats.CollectionDetails, destFormat *formats.Format, pe <-chan formats.ParseEvent) error {
	out, err := destFormat.NewCollection(*cd)
	if err != nil {
		return fmt.Errorf("unable to create a new collection in the %s format: %w", destFormat.Name, err)
	}
	defer out.Close()

	bar := progressbar.NewOptions(-1, progressbar.OptionFullWidth())

	for e := range pe {
		if e.N != 0 {
			bar.ChangeMax(e.N)
		}
		bar.Add(e.I)

		if e.Err != nil {
			progressbar.Bprintf(bar, "⛔️ Couldn't parse: %v\n", e.Err)
		} else if e.Recipe != nil {
			progressbar.Bprintf(bar, "📖 Found \"%s\"…\n", e.Recipe.Name())
			if err := out.Add(e.Recipe); err != nil {
				progressbar.Bprintf(bar, "  ⛔️ Writing error: %v\n", e.Err)
			}
			progressbar.Bprintf(bar, "  📥 …added into %s\n\n", out.Filename())
		}
	}

	return bar.Finish()
}

func makeRecipes(filename string, destFormat *formats.Format, pe <-chan formats.ParseEvent) error {
	return fmt.Errorf("single recipe output not yet implemented")
}
