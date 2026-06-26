package mela_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/mela"
)

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
	if !reflect.DeepEqual(ir.Ingredients, wantIngredients) {
		t.Errorf("Ingredients: want %v, got %v", wantIngredients, ir.Ingredients)
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
	if !reflect.DeepEqual(ir.Ingredients, want) {
		t.Errorf("Ingredients: want %#v, got %#v", want, ir.Ingredients)
	}
}
