package mela_test

import (
	"testing"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/mela"
)

func TestImportSourceAndDescription(t *testing.T) {
	ir := formats.NewInterchangeRecipe()
	ir.Title = "Abricotines"
	ir.Description = "Apricot Balls\n\nSeveral years ago I visited.\n\n\nThey were good."
	ir.Source = formats.Source{Name: "The Book of Jewish Food", URI: "urn:isbn:9780141928517"}

	got, err := mela.ImportRecipe(ir)
	if err != nil {
		t.Fatalf("ImportRecipe: %v", err)
	}
	mr, ok := got.(*mela.Recipe)
	if !ok {
		t.Fatalf("expected *mela.Recipe, got %T", got)
	}

	// Mela renders newlines literally, so paragraph breaks collapse to one.
	if want := "Apricot Balls\nSeveral years ago I visited.\nThey were good."; mr.Text != want {
		t.Errorf("Text = %q, want %q", mr.Text, want)
	}
	// The source URI becomes the Mela link.
	if mr.Link != "urn:isbn:9780141928517" {
		t.Errorf("Link = %q", mr.Link)
	}
	// Mela has no source field, so the source name is attributed in the notes.
	if mr.Notes != "From The Book of Jewish Food" {
		t.Errorf("Notes = %q", mr.Notes)
	}
}
