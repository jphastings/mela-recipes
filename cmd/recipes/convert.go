package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jphastings/recipes"
	"github.com/jphastings/recipes/internal/formats"
	cmdhelp "github.com/jphastings/recipes/internal/helpers"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert recipes between different formats",
	Long:  longExplain(),
	RunE: func(cmd *cobra.Command, args []string) error {
		network, err := cmd.Flags().GetBool("network")
		if err != nil {
			return err
		}
		o := withOwnershipPrompts(formats.ParseOptions{AllowNetwork: network})

		to, err := cmd.Flags().GetString("to")
		if err != nil {
			return err
		}
		filename, asType, destFormat := recipes.ParseDestination(to)

		pe, cd, err := recipes.ParseAll(args, o)
		if err != nil {
			return err
		}

		overwrite, err := cmd.Flags().GetBool("overwrite")
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
			cd.OverwriteExisting = overwrite
			return makeCollection(cd, destFormat, pe)
		} else {
			return makeRecipes(cmd, filename, destFormat, pe, overwrite)
		}
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().String("to", "", "The filename, extension, or format to convert into")
	convertCmd.MarkFlagRequired("to")
	convertCmd.Flags().Bool("network", false, "Allow network access while importing (eg. to fetch web-recipe images)")
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

	bar := progressbar.NewOptions(-1, progressbar.OptionFullWidth())

	i := 0
	for e := range pe {
		if e.N != 0 {
			bar.ChangeMax(e.N)
		}
		bar.Add(e.I)

		if e.Err != nil {
			progressbar.Bprintf(bar, "⛔️ Couldn't parse: %v\n", e.Err)
		} else if e.Recipe != nil {
			progressbar.Bprintf(bar, "📖 Found \"%s\"…\n", e.Recipe.Name())
			_, _ = e.Recipe.Standardize()

			if err := out.Add(e.Recipe); err != nil {
				progressbar.Bprintf(bar, "  ⛔️ Writing error: %v\n", err)
			}
			i++
			progressbar.Bprintf(bar, "  📥 …added into %s\n\n", out.Filename())
		}
	}

	// Close finalises the collection — for protected output this is where question
	// generation and encryption happen, so its error must surface (it's otherwise
	// the only signal of, eg., too few recipes to protect).
	if cerr := out.Close(); cerr != nil {
		return fmt.Errorf("unable to finish the %s collection: %w", destFormat.Name, cerr)
	}

	return bar.Finish()
}

func makeRecipes(cmd *cobra.Command, filename string, destFormat *formats.Format, pe <-chan formats.ParseEvent, overwrite bool) error {
	bar := progressbar.NewOptions(-1, progressbar.OptionFullWidth())

	for e := range pe {
		if e.N != 0 {
			bar.ChangeMax(e.N)
		}
		bar.Add(e.I)

		if e.Err != nil {
			progressbar.Bprintf(bar, "⛔️ Couldn't parse: %v\n", e.Err)
			continue
		}
		if e.Recipe == nil {
			continue
		}

		progressbar.Bprintf(bar, "📖 Found \"%s\"…\n", e.Recipe.Name())

		// Capture where the recipe came from before Standardize rewrites its
		// filename from the title, so --out-there can write back beside it.
		sourceDir := filepath.Dir(e.Recipe.Filename())
		_, _ = e.Recipe.Standardize()

		imported, err := destFormat.Import(e.Recipe)
		if err != nil {
			progressbar.Bprintf(bar, "  ⛔️ Conversion error: %v\n", err)
			continue
		}

		path, err := recipeOutputPath(cmd, filename, sourceDir, imported)
		if err != nil {
			return err
		}

		if err := writeRecipe(path, imported, overwrite); err != nil {
			progressbar.Bprintf(bar, "  ⛔️ Writing error: %v\n", err)
			continue
		}
		progressbar.Bprintf(bar, "  📥 …saved to %s\n\n", path)
	}

	return bar.Finish()
}

// recipeOutputPath decides where a converted recipe is written. An explicit
// destination filename (eg. --to dinner.md) is used verbatim; otherwise the
// recipe's filename is placed in the directory chosen by the --out-* flags
// (defaulting to the source directory).
func recipeOutputPath(cmd *cobra.Command, override, sourceDir string, r formats.Recipe) (string, error) {
	if override != "" {
		return override + r.Format().Extension, nil
	}

	outdir, err := cmdhelp.Outdir(cmd, sourceDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(outdir, filepath.Base(r.Filename())), nil
}

func writeRecipe(path string, r formats.Recipe, overwrite bool) (err error) {
	flags := os.O_CREATE | os.O_WRONLY
	if overwrite {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	f, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()

	return r.Marshal(f)
}
