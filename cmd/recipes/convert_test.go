package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/recipemd"
	"github.com/spf13/cobra"
)

// feed builds a closed ParseEvent channel carrying the given recipes.
func feed(recipes ...formats.Recipe) <-chan formats.ParseEvent {
	pe := make(chan formats.ParseEvent, len(recipes))
	for _, r := range recipes {
		pe <- formats.ParseEvent{Recipe: r, I: 1, N: len(recipes)}
	}
	close(pe)
	return pe
}

func testRecipe(title string) formats.InterchangeRecipe {
	ir := formats.NewInterchangeRecipe()
	ir.Title = title
	ir.Ingredients = []formats.TitledList{{List: []string{"1 egg"}}}
	ir.Instructions = []formats.TitledList{{List: []string{"Cook it."}}}
	return ir
}

func outDirCmd(t *testing.T, dir string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().String("out-dir", "", "")
	cmd.Flags().Bool("out-here", false, "")
	if err := cmd.Flags().Set("out-dir", dir); err != nil {
		t.Fatal(err)
	}
	return cmd
}

func TestMakeRecipesWritesEachRecipe(t *testing.T) {
	dir := t.TempDir()
	cmd := outDirCmd(t, dir)

	if err := makeRecipes(cmd, "", recipemd.FormatInfo, feed(testRecipe("Test Recipe")), false); err != nil {
		t.Fatalf("makeRecipes: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "test-recipe.md"))
	if err != nil {
		t.Fatalf("reading converted recipe: %v", err)
	}
	for _, want := range []string{"# Test Recipe", "- 1 egg", "1. Cook it."} {
		if !strings.Contains(string(data), want) {
			t.Errorf("output is missing %q:\n%s", want, data)
		}
	}
}

func TestMakeRecipesExplicitName(t *testing.T) {
	dir := t.TempDir()
	cmd := outDirCmd(t, dir)
	name := filepath.Join(dir, "dinner")

	if err := makeRecipes(cmd, name, recipemd.FormatInfo, feed(testRecipe("Anything")), false); err != nil {
		t.Fatalf("makeRecipes: %v", err)
	}
	if _, err := os.Stat(name + ".md"); err != nil {
		t.Errorf("expected %s.md to exist: %v", name, err)
	}
}

func TestMakeRecipesOverwriteGuard(t *testing.T) {
	dir := t.TempDir()
	cmd := outDirCmd(t, dir)
	out := filepath.Join(dir, "test-recipe.md")

	if err := os.WriteFile(out, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	// Without --overwrite the existing file is left untouched.
	if err := makeRecipes(cmd, "", recipemd.FormatInfo, feed(testRecipe("Test Recipe")), false); err != nil {
		t.Fatalf("makeRecipes: %v", err)
	}
	if data, _ := os.ReadFile(out); string(data) != "existing" {
		t.Errorf("file was overwritten without --overwrite: %q", data)
	}

	// With --overwrite it is replaced.
	if err := makeRecipes(cmd, "", recipemd.FormatInfo, feed(testRecipe("Test Recipe")), true); err != nil {
		t.Fatalf("makeRecipes (overwrite): %v", err)
	}
	if data, _ := os.ReadFile(out); !strings.Contains(string(data), "# Test Recipe") {
		t.Errorf("file was not overwritten with --overwrite: %q", data)
	}
}
