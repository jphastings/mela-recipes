package mela

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/protected"
	"github.com/jphastings/recipes/utils"
)

func buildRecipes(t *testing.T, n int) []*Recipe {
	t.Helper()
	words := []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet"}

	recipes := make([]*Recipe, n)
	for i := 0; i < n; i++ {
		r := &Recipe{
			Title: fmt.Sprintf("Recipe number %s", words[i%len(words)]),
			Text:  "This tasty dish brings comfort. Serve warm with extra sauce.",
			Instructions: SectionedSequence(
				"Gather every needed ingredient together carefully\n" +
					"Combine flour sugar butter together slowly\n" +
					"Bake until golden brown throughout completely",
			),
		}
		if err := r.SetBook("9781234567897", utils.MustParsePages(fmt.Sprintf("%d", 40+i)), uint(i%3+1)); err != nil {
			t.Fatalf("SetBook: %v", err)
		}
		recipes[i] = r
	}
	return recipes
}

// writeProtected writes the recipes to a .protectedrecipes file and returns its path.
func writeProtected(t *testing.T, recipes []*Recipe) string {
	t.Helper()
	cd := formats.CollectionDetails{Filename: filepath.Join(t.TempDir(), "book")}
	w, err := NewProtectedCollection(cd)
	if err != nil {
		t.Fatalf("NewProtectedCollection: %v", err)
	}
	for _, r := range recipes {
		if err := w.Add(r); err != nil {
			t.Fatalf("Add: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return w.Filename()
}

func TestParseProtectedRequiresPrompts(t *testing.T) {
	path := writeProtected(t, buildRecipes(t, 8))

	_, _, err := ParseProtected(formats.Bundle{path}, formats.ParseOptions{})
	if err == nil || !strings.Contains(err.Error(), "interactive ownership") {
		t.Fatalf("got %v, want a missing-prompts error", err)
	}
}

func TestParseProtectedWrongAnswers(t *testing.T) {
	path := writeProtected(t, buildRecipes(t, 8))

	o := formats.ParseOptions{
		AskOwnership: func(string) (string, error) { return "wrong", nil },
	}
	_, _, err := ParseProtected(formats.Bundle{path}, o)
	if !errors.Is(err, protected.ErrIncorrectAnswers) {
		t.Fatalf("got %v, want protected.ErrIncorrectAnswers", err)
	}
}
