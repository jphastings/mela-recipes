package recipeml

import (
	"path/filepath"
	"testing"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/ingredients"
)

// drain collects every recipe a fixture parses into, failing on any parse error.
func drain(t *testing.T, name string) []formats.InterchangeRecipe {
	t.Helper()
	pe, cd, err := Parse(formats.Bundle{filepath.Join("fixtures", name)}, formats.ParseOptions{})
	if err != nil {
		t.Fatalf("Parse(%s): %v", name, err)
	}
	if cd != nil {
		t.Errorf("expected nil CollectionDetails, got %+v", cd)
	}

	var recipes []formats.InterchangeRecipe
	for e := range pe {
		if e.Err != nil {
			t.Fatalf("parse event error: %v", e.Err)
		}
		if e.Recipe != nil {
			ir, err := e.Recipe.Export()
			if err != nil {
				t.Fatalf("Export: %v", err)
			}
			recipes = append(recipes, ir)
		}
	}
	return recipes
}

func lines(g formats.IngredientGroup) []string {
	var out []string
	for _, iu := range g.Items {
		out = append(out, ingredients.FormatIngredientUse(iu))
	}
	return out
}

func TestParseSingle(t *testing.T) {
	recipes := drain(t, "single.xml")
	if len(recipes) != 1 {
		t.Fatalf("got %d recipes, want 1", len(recipes))
	}
	r := recipes[0]

	if r.Title != "Chocolate Chip Cookies" {
		t.Errorf("Title = %q", r.Title)
	}
	if r.Description != "Classic cookies." {
		t.Errorf("Description = %q", r.Description)
	}
	if r.Notes != "Best fresh." {
		t.Errorf("Notes = %q", r.Notes)
	}
	if want := []string{"Dessert", "Cookies"}; !equal(r.Tags, want) {
		t.Errorf("Tags = %v, want %v", r.Tags, want)
	}
	if r.Yield != "24 cookies" {
		t.Errorf("Yield = %q", r.Yield)
	}

	if len(r.Ingredients) != 1 {
		t.Fatalf("ingredient groups = %d, want 1", len(r.Ingredients))
	}
	want := []string{"2 cup flour", "½ tsp salt (sifted)", "1 cup butter (softened)", "2 eggs"}
	if got := lines(r.Ingredients[0]); !equal(got, want) {
		t.Errorf("ingredients = %#v, want %#v", got, want)
	}

	if len(r.Instructions) != 1 {
		t.Fatalf("instruction groups = %d, want 1", len(r.Instructions))
	}
	wantSteps := []string{"Cream the butter and sugar.", "Bake at 375F for 10 minutes."}
	if got := r.Instructions[0].List; !equal(got, wantSteps) {
		t.Errorf("instructions = %#v, want %#v (inline <temp> markup should flatten)", got, wantSteps)
	}
}

func TestParseMulti(t *testing.T) {
	recipes := drain(t, "multi.xml")
	if len(recipes) != 2 {
		t.Fatalf("got %d recipes, want 2", len(recipes))
	}
	if recipes[0].Title != "Pancakes" || recipes[1].Title != "Lemonade" {
		t.Errorf("titles = %q, %q", recipes[0].Title, recipes[1].Title)
	}
	if got := lines(recipes[1].Ingredients[0]); !equal(got, []string{"4 lemons", "1 cup sugar"}) {
		t.Errorf("Lemonade ingredients = %#v", got)
	}
}

func TestParseSections(t *testing.T) {
	recipes := drain(t, "sections.xml")
	if len(recipes) != 1 {
		t.Fatalf("got %d recipes, want 1", len(recipes))
	}
	r := recipes[0]

	if len(r.Ingredients) != 2 {
		t.Fatalf("ingredient groups = %d, want 2", len(r.Ingredients))
	}
	if r.Ingredients[0].Title != "Sponge" || r.Ingredients[1].Title != "Icing" {
		t.Errorf("group titles = %q, %q", r.Ingredients[0].Title, r.Ingredients[1].Title)
	}
	if got := lines(r.Ingredients[1]); !equal(got, []string{"2 cup icing sugar"}) {
		t.Errorf("Icing ingredients = %#v", got)
	}
}

func TestRejectsNonRecipeML(t *testing.T) {
	pe, _, err := Parse(formats.Bundle{filepath.Join("fixtures", "notrecipe.xml")}, formats.ParseOptions{})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	var sawError bool
	for e := range pe {
		if e.Err != nil {
			sawError = true
		}
		if e.Recipe != nil {
			t.Errorf("expected no recipe from non-RecipeML XML, got %q", e.Recipe.Name())
		}
	}
	if !sawError {
		t.Error("expected a non-fatal error for non-RecipeML XML")
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
