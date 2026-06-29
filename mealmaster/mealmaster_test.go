package mealmaster

import (
	"math/big"
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
		t.Errorf("expected nil CollectionDetails for a flat .mmf, got %+v", cd)
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

// lines renders an ingredient group's items back to readable strings.
func lines(g formats.IngredientGroup) []string {
	var out []string
	for _, iu := range g.Items {
		out = append(out, ingredients.FormatIngredientUse(iu))
	}
	return out
}

func TestParseSingle(t *testing.T) {
	recipes := drain(t, "single.mmf")
	if len(recipes) != 1 {
		t.Fatalf("got %d recipes, want 1", len(recipes))
	}
	r := recipes[0]

	if r.Title != "Chocolate Chip Cookies" {
		t.Errorf("Title = %q", r.Title)
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
	want := []string{"2 cup Flour", "½ tsp Salt", "1 cup Butter (softened)", "2 Eggs chilled"}
	if got := lines(r.Ingredients[0]); !equal(got, want) {
		t.Errorf("ingredients = %#v, want %#v", got, want)
	}

	if len(r.Instructions) != 1 {
		t.Fatalf("instruction groups = %d, want 1", len(r.Instructions))
	}
	wantSteps := []string{
		"Cream butter and sugar.  Add eggs and beat well.",
		"Bake at 375F for 10 minutes.",
	}
	if got := r.Instructions[0].List; !equal(got, wantSteps) {
		t.Errorf("instructions = %#v, want %#v", got, wantSteps)
	}
}

func TestParseMulti(t *testing.T) {
	recipes := drain(t, "multi.mmf")
	if len(recipes) != 2 {
		t.Fatalf("got %d recipes, want 2", len(recipes))
	}
	if recipes[0].Title != "Pancakes" || recipes[1].Title != "Lemonade" {
		t.Errorf("titles = %q, %q", recipes[0].Title, recipes[1].Title)
	}
	if got := lines(recipes[1].Ingredients[0]); !equal(got, []string{"4 Lemons", "1 cup Sugar"}) {
		t.Errorf("Lemonade ingredients = %#v", got)
	}
}

func TestParseTwoColumn(t *testing.T) {
	recipes := drain(t, "twocolumn.mmf")
	if len(recipes) != 1 {
		t.Fatalf("got %d recipes, want 1", len(recipes))
	}
	r := recipes[0]

	if len(r.Ingredients) != 2 {
		t.Fatalf("ingredient groups = %d, want 2 (base + Icing section)", len(r.Ingredients))
	}
	base := lines(r.Ingredients[0])
	if want := []string{"2 cup Flour", "1 tsp Salt", "1 cup Sugar", "3 Eggs"}; !equal(base, want) {
		t.Errorf("base ingredients = %#v, want %#v (two columns should interleave row by row)", base, want)
	}
	if r.Ingredients[1].Title != "Icing" {
		t.Errorf("second group title = %q, want Icing", r.Ingredients[1].Title)
	}
}

func TestExpandUnit(t *testing.T) {
	cases := map[string]string{"c": "cup", "ts": "tsp", "T": "tbsp", "ea": "", "cn": "can", "zz": "zz"}
	for in, want := range cases {
		if got := expandUnit(in); got != want {
			t.Errorf("expandUnit(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestExactFraction confirms the common ASCII fractions reach the parser exactly,
// not as a lossy decimal.
func TestExactFraction(t *testing.T) {
	iu := buildIngredient("1/3", "c", "milk", 0)
	want := big.NewRat(1, 3)
	if iu.Quantity.Amount == nil || (*big.Rat)(iu.Quantity.Amount).Cmp(want) != 0 {
		t.Errorf("amount = %v, want 1/3", iu.Quantity.Amount)
	}
	if iu.Quantity.Type != ingredients.UnitCup {
		t.Errorf("unit = %s, want CUP", iu.Quantity.Type)
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
