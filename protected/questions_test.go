package protected

import (
	"strings"
	"testing"

	"github.com/jphastings/recipes/internal/formats"
)

func TestLocatorRevealsAnswer(t *testing.T) {
	titleLoc := "Look at the recipe titled 'Zeera Aloo'. "
	pageLoc := "Look at the first recipe on page 81. "

	cases := []struct {
		name    string
		locator string
		answer  string
		want    bool
	}{
		{"title is the whole answer", titleLoc, "Zeera Aloo", true},
		{"a title word is the answer", titleLoc, "zeera", true},
		{"case/space insensitive", titleLoc, "  ZEERA  ", true},
		{"unrelated word with title locator", titleLoc, "cumin", false},
		{"page locator hides content word", pageLoc, "flambé", false},
		{"empty answer", titleLoc, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := locatorRevealsAnswer(c.locator, c.answer); got != c.want {
				t.Errorf("locatorRevealsAnswer(%q, %q) = %v, want %v", c.locator, c.answer, got, c.want)
			}
		})
	}
}

// TestTitleLocatedRecipesNeverAskTheTitle reproduces the reported flaw: recipes
// without a book reference are located by title, so the "what is the title?"
// question would give away its own answer and must never be produced.
func TestTitleLocatedRecipesNeverAskTheTitle(t *testing.T) {
	titles := []string{
		"Zeera Aloo", "Aloo Bharta", "Gajjar Halwa", "Sooji Ladoo",
		"Tehri Rice", "Courgette Sabzi", "Lachedar Paratha", "Methi Murgh",
	}
	var recipes []formats.InterchangeRecipe
	for _, title := range titles {
		ir := formats.NewInterchangeRecipe() // no ID → no book ref → located by title
		ir.Title = title
		ir.Description = "This warming dish brings welcome comfort. Serve piping with fresh flatbread."
		ir.Instructions = []formats.TitledList{{List: []string{
			"Gather every needed ingredient together",
			"Simmer everything gently until softened",
			"Garnish generously before serving warm",
		}}}
		recipes = append(recipes, ir)
	}

	// Randomised generation — run enough rounds to exercise the maker ordering.
	for i := 0; i < 100; i++ {
		qs, as, err := prepareQuestions(recipes, defaultQuestionCount)
		if err != nil {
			t.Fatalf("prepareQuestions: %v", err)
		}
		for j, q := range qs {
			if strings.Contains(q, "What is the recipe's title?") {
				t.Fatalf("title question produced for a title-located recipe: %q", q)
			}
			if locatorRevealsAnswer(q, as[j]) {
				t.Errorf("answer %q is revealed within its own question %q", as[j], q)
			}
		}
	}
}
