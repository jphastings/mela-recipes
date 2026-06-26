package protected

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/mela"
	"github.com/jphastings/recipes/utils"
	"github.com/yeka/zip"
)

// makeRecipes builds n distinct, fully-populated Mela recipes, each carrying a
// book reference so question generation can produce page-based locators.
func makeRecipes(t *testing.T, n int) []formats.Recipe {
	t.Helper()
	words := []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet"}

	recipes := make([]formats.Recipe, n)
	for i := 0; i < n; i++ {
		r := &mela.Recipe{
			Title: fmt.Sprintf("Recipe number %s", words[i%len(words)]),
			Text:  fmt.Sprintf("This %s dish brings warmth. Serve immediately while piping hot.", words[i%len(words)]),
			Instructions: mela.SectionedSequence(
				"Gather every needed ingredient together carefully\n" +
					fmt.Sprintf("Combine flour sugar butter %s slowly\n", words[i%len(words)]) +
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

func interchanges(t *testing.T, recipes []formats.Recipe) []formats.InterchangeRecipe {
	t.Helper()
	irs := make([]formats.InterchangeRecipe, len(recipes))
	for i, r := range recipes {
		ir, err := r.Export()
		if err != nil {
			t.Fatalf("Export: %v", err)
		}
		irs[i] = ir
	}
	return irs
}

// TestMelaRoundTrip drives the full pipeline: question generation from real Mela
// recipes, encryption, then decryption with correct answers and re-parsing.
func TestMelaRoundTrip(t *testing.T) {
	recipes := makeRecipes(t, defaultQuestionCount)

	// Generate the questions once and keep the answers, so the read-side callback
	// can answer them by exact question text.
	questions, answers, err := prepareQuestions(interchanges(t, recipes), defaultQuestionCount)
	if err != nil {
		t.Fatalf("prepareQuestions: %v", err)
	}
	answerByQuestion := map[string]string{}
	for i, q := range questions {
		answerByQuestion[q] = answers[i]
	}

	entries := make([]archiveEntry, len(recipes))
	wantTitles := map[string]string{}
	for i, r := range recipes {
		entries[i] = archiveEntry{name: r.Filename(), write: r.Marshal}
		wantTitles[r.Filename()] = r.Name()
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if err := writeArchive(zw, entries, questions, answers, defaultRequiredCorrect); err != nil {
		t.Fatalf("writeArchive: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}

	r := bytes.NewReader(buf.Bytes())
	decrypted, err := Read(r, r.Size(), answerFunc(answerByQuestion), noExplain)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(decrypted) != len(recipes) {
		t.Fatalf("got %d entries, want %d", len(decrypted), len(recipes))
	}

	for _, e := range decrypted {
		rc, err := e.Open()
		if err != nil {
			t.Fatalf("open %q: %v", e.Name, err)
		}
		parsed, err := mela.ParseRecipeStream(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("parse %q: %v", e.Name, err)
		}
		if parsed.Title != wantTitles[e.Name] {
			t.Errorf("%q title = %q, want %q", e.Name, parsed.Title, wantTitles[e.Name])
		}
	}
}

// TestCreateOpenWrongAnswers exercises the file-based Create/Open convenience and
// confirms that incorrect answers are rejected.
func TestCreateOpenWrongAnswers(t *testing.T) {
	recipes := makeRecipes(t, defaultQuestionCount)
	path := filepath.Join(t.TempDir(), "book.protectedrecipes")

	if err := Create(path, recipes); err != nil {
		t.Fatalf("Create: %v", err)
	}

	wrong := func(string) (string, error) { return "definitely wrong", nil }
	_, closer, err := Open(path, wrong, noExplain)
	if err != ErrIncorrectAnswers {
		if closer != nil {
			closer.Close()
		}
		t.Fatalf("got %v, want ErrIncorrectAnswers", err)
	}
}
