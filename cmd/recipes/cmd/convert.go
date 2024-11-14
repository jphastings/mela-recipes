/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/jphastings/recipes"
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert recipes between different formats",
	Long:  fmt.Sprintf(`Converts cookbook recipes between supported formats: %s`, recipes.AvailableFormatsList()),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("convert command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	// Todo: https://stackoverflow.com/questions/50824554/permitted-flag-values-for-cobra
}
