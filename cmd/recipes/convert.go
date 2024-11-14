package main

import (
	"fmt"
	"strings"

	"github.com/jphastings/recipes"
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert recipes between different formats",
	Long:  longExplain(),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("convert command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().String("to", "", "Destination filename, extension, or format")
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
