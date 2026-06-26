package mela_test

import (
	"fmt"
	"strings"

	"github.com/jphastings/recipes/mela"
)

// ParseRecipeStream decodes a single recipe from any io.Reader.
func ExampleParseRecipeStream() {
	r, err := mela.ParseRecipeStream(strings.NewReader(
		`{"title":"Pancakes","yield":"4","ingredients":"Flour","instructions":"Mix"}`,
	))
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(r.Name())
	// Output: Pancakes
}

// ParseRecipesFile streams the recipes from a .melarecipes archive as they are
// decoded. Each event carries either a recipe, an error, or progress counts.
func ExampleParseRecipesFile() {
	events, _, err := mela.ParseRecipesFile("fixtures/a+b.melarecipes")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	for e := range events {
		if e.Recipe != nil {
			fmt.Println(e.Recipe.Name())
		}
	}
	// Output:
	// B title
	// A title
}
