package mela_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/ingredients"
	"github.com/jphastings/recipes/mela"
)

// ingredientLines renders structured ingredient groups back to their canonical
// string lines, so tests can assert on readable text without the per-item UUIDs.
func ingredientLines(groups []formats.IngredientGroup) []formats.TitledList {
	out := make([]formats.TitledList, 0, len(groups))
	for _, g := range groups {
		tl := formats.TitledList{Title: g.Title}
		for _, iu := range g.Items {
			tl.List = append(tl.List, ingredients.FormatIngredientUse(iu))
		}
		out = append(out, tl)
	}
	return out
}

func TestExport(t *testing.T) {
	events, err := mela.ParseRecipeFile("fixtures/a.melarecipe")
	if err != nil {
		t.Fatal(err)
	}
	r := drain(t, events)[0]

	ir, err := r.Export()
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	if ir.Title != "A title" {
		t.Errorf("Title: want %q, got %q", "A title", ir.Title)
	}
	if ir.Description != "A text" {
		t.Errorf("Description: want %q, got %q", "A text", ir.Description)
	}
	if ir.Yield != "1" {
		t.Errorf("Yield: want %q, got %q", "1", ir.Yield)
	}
	if want := []string{"a", "aa", "aaa"}; !reflect.DeepEqual(ir.Tags, want) {
		t.Errorf("Tags: want %v, got %v", want, ir.Tags)
	}

	wantIngredients := []formats.TitledList{{Title: "", List: []string{"A ingredients"}}}
	if got := ingredientLines(ir.Ingredients); !reflect.DeepEqual(got, wantIngredients) {
		t.Errorf("Ingredients: want %v, got %v", wantIngredients, got)
	}

	if ir.PrepTime == nil || *ir.PrepTime != time.Minute {
		t.Errorf("PrepTime: want 1m, got %v", ir.PrepTime)
	}
	if ir.CookTime == nil || *ir.CookTime != time.Hour {
		t.Errorf("CookTime: want 1h, got %v", ir.CookTime)
	}
	if ir.TotalTime != nil {
		t.Errorf("TotalTime: want nil, got %v", ir.TotalTime)
	}
}

// TestExportSections covers the non-trivial part of Export: splitting a
// SectionedSequence back into ordered, cleanly-titled lists.
func TestExportSections(t *testing.T) {
	r := &mela.Recipe{
		Ingredients: mela.SectionedSequence("Base item\n# Sauce\nButter\nFlour"),
	}

	ir, err := r.Export()
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	want := []formats.TitledList{
		{Title: "", List: []string{"Base item"}},
		{Title: "Sauce", List: []string{"Butter", "Flour"}},
	}
	if got := ingredientLines(ir.Ingredients); !reflect.DeepEqual(got, want) {
		t.Errorf("Ingredients: want %#v, got %#v", want, got)
	}
}
