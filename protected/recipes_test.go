package protected

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"testing"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/standardize"
	"github.com/jphastings/recipes/utils"
	"github.com/yeka/zip"
)

// fakeRecipe is a minimal formats.Recipe used to exercise protected without
// depending on any concrete recipe format (protected is format-agnostic).
type fakeRecipe struct {
	title string
	id    string
	body  string
}

func (r *fakeRecipe) Name() string                              { return r.title }
func (r *fakeRecipe) Format() *formats.Format                   { return nil }
func (r *fakeRecipe) Filename() string                          { return r.title + ".melarecipe" }
func (r *fakeRecipe) Standardize() ([]standardize.Std, error)   { return nil, nil }
func (r *fakeRecipe) Marshal(w io.Writer) error                 { _, err := io.WriteString(w, r.body); return err }
func (r *fakeRecipe) Export() (formats.InterchangeRecipe, error) {
	ir := formats.NewInterchangeRecipe()
	ir.ID = r.id
	ir.Title = r.title
	ir.Description = "This tasty dish brings comfort. Serve warm with extra sauce."
	ir.Instructions = []formats.TitledList{{Title: "", List: []string{
		"Gather every needed ingredient together carefully",
		"Combine flour sugar butter together slowly",
		"Bake until golden brown throughout completely",
	}}}
	return ir, nil
}

// makeRecipes builds n distinct recipes, each with a book reference so question
// generation can produce page-based locators.
func makeRecipes(n int) []formats.Recipe {
	words := []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet"}

	recipes := make([]formats.Recipe, n)
	for i := 0; i < n; i++ {
		title := fmt.Sprintf("Recipe number %s", words[i%len(words)])
		br := utils.BookRef{
			Book:         utils.Book{ISBN13: "9781234567897"},
			Pages:        utils.MustParsePages(fmt.Sprintf("%d", 40+i)),
			RecipeNumber: uint(i%3 + 1),
		}
		recipes[i] = &fakeRecipe{title: title, id: br.URN(), body: "BODY:" + title}
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

// TestRecipeRoundTrip drives the full pipeline: question generation from recipes,
// encryption, then decryption with correct answers, checking each entry's bytes.
func TestRecipeRoundTrip(t *testing.T) {
	recipes := makeRecipes(defaultQuestionCount)

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
	wantBody := map[string]string{}
	for i, r := range recipes {
		entries[i] = archiveEntry{name: r.Filename(), write: r.Marshal}
		wantBody[r.Filename()] = "BODY:" + r.Name()
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
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("read %q: %v", e.Name, err)
		}
		if string(data) != wantBody[e.Name] {
			t.Errorf("%q = %q, want %q", e.Name, data, wantBody[e.Name])
		}
	}
}

// TestCreateOpenWrongAnswers exercises the file-based Create/Open convenience and
// confirms that incorrect answers are rejected.
func TestCreateOpenWrongAnswers(t *testing.T) {
	recipes := makeRecipes(defaultQuestionCount)
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
