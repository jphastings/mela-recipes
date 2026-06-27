package recipemd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/jphastings/recipes/internal/formats"
)

func loadFixture(t *testing.T, name string) formats.InterchangeRecipe {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("fixtures", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	ir, err := parseRecipe(data)
	if err != nil {
		t.Fatalf("parseRecipe(%s): %v", name, err)
	}
	return ir
}

func sections(titled ...formats.TitledList) []formats.TitledList { return titled }
func sec(title string, items ...string) formats.TitledList {
	return formats.TitledList{Title: title, List: items}
}

func TestParseSimple(t *testing.T) {
	ir := loadFixture(t, "simple.md")

	if ir.Title != "Pancakes" {
		t.Errorf("Title = %q", ir.Title)
	}
	if ir.Description != "Light and fluffy weekend pancakes." {
		t.Errorf("Description = %q", ir.Description)
	}
	if want := []string{"breakfast", "easy"}; !reflect.DeepEqual(ir.Tags, want) {
		t.Errorf("Tags = %v, want %v", ir.Tags, want)
	}
	if ir.Yield != "12 pancakes" {
		t.Errorf("Yield = %q", ir.Yield)
	}

	wantIng := sections(sec("", "2 cups flour", "2 eggs", "1 1/2 cups milk"))
	if !reflect.DeepEqual(ir.Ingredients, wantIng) {
		t.Errorf("Ingredients = %#v", ir.Ingredients)
	}
	wantSteps := sections(sec("", "Mix the dry ingredients.", "Whisk in the eggs and milk.", "Cook on a hot, greased griddle."))
	if !reflect.DeepEqual(ir.Instructions, wantSteps) {
		t.Errorf("Instructions = %#v", ir.Instructions)
	}
}

func TestParseGrouped(t *testing.T) {
	ir := loadFixture(t, "grouped.md")

	if ir.Title != "Trifle" {
		t.Errorf("Title = %q", ir.Title)
	}
	if ir.Description != "" {
		t.Errorf("Description = %q, want empty", ir.Description)
	}
	if ir.Yield != "8 servings" {
		t.Errorf("Yield = %q", ir.Yield)
	}

	wantIng := sections(
		sec("", "200 g sponge fingers"),
		sec("Custard", "500 ml milk", "4 egg yolks"),
	)
	if !reflect.DeepEqual(ir.Ingredients, wantIng) {
		t.Errorf("Ingredients = %#v", ir.Ingredients)
	}
	wantSteps := sections(
		sec("Custard", "Heat the milk until steaming.", "Whisk in the yolks and thicken."),
		sec("Assembly", "Layer the sponge and custard.", "Chill before serving."),
	)
	if !reflect.DeepEqual(ir.Instructions, wantSteps) {
		t.Errorf("Instructions = %#v", ir.Instructions)
	}
}

func TestRoundTrip(t *testing.T) {
	ir := formats.NewInterchangeRecipe()
	ir.Title = "Round Trip"
	ir.Description = "A small test recipe."
	ir.Tags = []string{"test", "sample"}
	ir.Yield = "4 servings"
	ir.Ingredients = sections(
		sec("", "2 cups flour", "1 egg"),
		sec("Sauce", "200 ml cream"),
	)
	ir.Instructions = sections(
		sec("", "Mix everything together.", "Bake until golden."),
		sec("Finishing", "Dust with sugar."),
	)

	// Import -> Marshal -> Parse should return the same interchange fields.
	rec, err := Import(ir)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	var buf bytes.Buffer
	if err := rec.Marshal(&buf); err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	got, err := parseRecipe(buf.Bytes())
	if err != nil {
		t.Fatalf("parseRecipe after Marshal: %v\n---\n%s", err, buf.String())
	}

	for _, c := range []struct {
		name      string
		got, want any
	}{
		{"Title", got.Title, ir.Title},
		{"Description", got.Description, ir.Description},
		{"Tags", got.Tags, ir.Tags},
		{"Yield", got.Yield, ir.Yield},
		{"Ingredients", got.Ingredients, ir.Ingredients},
		{"Instructions", got.Instructions, ir.Instructions},
	} {
		if !reflect.DeepEqual(c.got, c.want) {
			t.Errorf("%s round-tripped to %#v, want %#v\n--- markdown ---\n%s", c.name, c.got, c.want, buf.String())
		}
	}
}

func TestParseRejectsNonRecipeMD(t *testing.T) {
	cases := map[string]string{
		"no rules": "# Just a heading\n\nSome prose, no recipe structure.\n",
		"one rule": "# Title\n\nDesc\n\n---\n\n- an ingredient\n",
		"no title": "Some text.\n\n---\n\n- item\n\n---\n\n1. step\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parseRecipe([]byte(src)); !errors.Is(err, errNotRecipeMD) {
				t.Errorf("got %v, want errNotRecipeMD", err)
			}
		})
	}
}

func TestParseEmitsRecipe(t *testing.T) {
	pe, cd, err := Parse(formats.Bundle{filepath.Join("fixtures", "simple.md")}, formats.ParseOptions{})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cd != nil {
		t.Errorf("expected nil CollectionDetails, got %+v", cd)
	}

	var got formats.Recipe
	for e := range pe {
		if e.Err != nil {
			t.Fatalf("ParseEvent error: %v", e.Err)
		}
		if e.Recipe != nil {
			got = e.Recipe
		}
	}
	if got == nil || got.Name() != "Pancakes" {
		t.Fatalf("got %v", got)
	}
	if fn := got.Filename(); !strings.HasSuffix(fn, ".md") {
		t.Errorf("Filename() = %q, want a .md name", fn)
	}
}
