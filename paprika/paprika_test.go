package paprika_test

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/paprika"
	"github.com/jphastings/recipes/utils"
)

// onePixelPNG returns the bytes of a 1×1 white PNG, used as recipe photo data.
func onePixelPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func recipeJSON(t *testing.T) string {
	t.Helper()
	photo := base64.StdEncoding.EncodeToString(onePixelPNG(t))
	return `{"uid":"a","name":"A title","description":"A description",` +
		`"ingredients":"A ingredients","directions":"A directions","notes":"A notes",` +
		`"servings":"4","prep_time":"20 mins.","cook_time":"1 hour",` +
		`"categories":["dinner","quick"],"photo_data":"` + photo + `"}`
}

// writeRecipeFile writes a gzip-compressed .paprikarecipe file and returns its path.
func writeRecipeFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	if _, err := gz.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

// writeCollection writes a .paprikarecipes zip whose entries are gzipped JSON.
func writeCollection(t *testing.T, dir, name string, entries map[string]string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for entry, body := range entries {
		w, err := zw.Create(entry)
		if err != nil {
			t.Fatal(err)
		}
		gz := gzip.NewWriter(w)
		if _, err := gz.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
		if err := gz.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func drain(t *testing.T, events <-chan formats.ParseEvent) []*paprika.Recipe {
	t.Helper()
	var recipes []*paprika.Recipe
	for e := range events {
		if e.Err != nil {
			t.Fatalf("unexpected parse error: %v", e.Err)
		}
		if e.Recipe == nil {
			continue
		}
		r, ok := e.Recipe.(*paprika.Recipe)
		if !ok {
			t.Fatalf("expected *paprika.Recipe, got %T", e.Recipe)
		}
		recipes = append(recipes, r)
	}
	return recipes
}

func TestParseRecipeFile(t *testing.T) {
	dir := t.TempDir()
	path := writeRecipeFile(t, dir, "a.paprikarecipe", recipeJSON(t))

	events, err := paprika.ParseRecipeFile(path)
	if err != nil {
		t.Fatal(err)
	}
	recipes := drain(t, events)
	if len(recipes) != 1 {
		t.Fatalf("expected 1 recipe, got %d", len(recipes))
	}
	r := recipes[0]

	if r.Name() != "A title" {
		t.Errorf("Name: want %q, got %q", "A title", r.Name())
	}
	if r.Servings != "4" {
		t.Errorf("Servings: want %q, got %q", "4", r.Servings)
	}
	if want := map[string][]string{"": {"A ingredients"}}; !reflect.DeepEqual(r.Ingredients.Parse(), want) {
		t.Errorf("Ingredients: want %v, got %v", want, r.Ingredients.Parse())
	}
	if want := []string{"dinner", "quick"}; !reflect.DeepEqual(r.Categories, want) {
		t.Errorf("Categories: want %v, got %v", want, r.Categories)
	}
	if prep, _ := r.PrepTime.Parse(); prep == nil || *prep != 20*time.Minute {
		t.Errorf("PrepTime: want 20m, got %v", prep)
	}
	if cook, _ := r.CookTime.Parse(); cook == nil || *cook != time.Hour {
		t.Errorf("CookTime: want 1h, got %v", cook)
	}
	if len(r.PhotoData) == 0 {
		t.Error("PhotoData: want photo bytes, got none")
	}
}

// TestCollectionRoundTrip exercises the full gzip-in-zip pipeline: read a
// .paprikarecipes archive, write it back out, and read it again.
func TestCollectionRoundTrip(t *testing.T) {
	dir := t.TempDir()
	collection := writeCollection(t, dir, "book.paprikarecipes", map[string]string{
		"a.paprikarecipe": recipeJSON(t),
		"b.paprikarecipe": `{"uid":"b","name":"B title","ingredients":"B ing","directions":"B dir"}`,
	})

	events, _, err := paprika.ParseRecipesFile(collection)
	if err != nil {
		t.Fatal(err)
	}
	originals := drain(t, events)
	if len(originals) != 2 {
		t.Fatalf("expected 2 recipes, got %d", len(originals))
	}

	out := filepath.Join(dir, "round-trip")
	cw, err := paprika.NewCollection(formats.CollectionDetails{Filename: out, OverwriteExisting: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range originals {
		if err := cw.Add(r); err != nil {
			t.Fatal(err)
		}
	}
	if err := cw.Close(); err != nil {
		t.Fatal(err)
	}

	events, _, err = paprika.ParseRecipesFile(out + ".paprikarecipes")
	if err != nil {
		t.Fatal(err)
	}
	got := drain(t, events)

	gotTitles := map[string]bool{}
	for _, r := range got {
		gotTitles[r.Name()] = true
	}
	for _, want := range []string{"A title", "B title"} {
		if !gotTitles[want] {
			t.Errorf("round-trip lost recipe %q (got %v)", want, gotTitles)
		}
	}
}

// TestMarshalIsGzip confirms a single recipe marshals to gzip-compressed JSON
// that parses back to an equivalent recipe.
func TestMarshalIsGzip(t *testing.T) {
	dir := t.TempDir()
	events, err := paprika.ParseRecipeFile(writeRecipeFile(t, dir, "a.paprikarecipe", recipeJSON(t)))
	if err != nil {
		t.Fatal(err)
	}
	r := drain(t, events)[0]

	var buf bytes.Buffer
	if err := r.Marshal(&buf); err != nil {
		t.Fatal(err)
	}
	if _, err := gzip.NewReader(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("Marshal output is not gzip: %v", err)
	}

	reparsed, err := paprika.ParseRecipeStream(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if reparsed.Name() != "A title" {
		t.Errorf("reparsed Name: want %q, got %q", "A title", reparsed.Name())
	}
}

func TestExport(t *testing.T) {
	dir := t.TempDir()
	events, err := paprika.ParseRecipeFile(writeRecipeFile(t, dir, "a.paprikarecipe", recipeJSON(t)))
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
	if ir.Description != "A description" {
		t.Errorf("Description: want %q, got %q", "A description", ir.Description)
	}
	if ir.Yield != "4" {
		t.Errorf("Yield: want %q, got %q", "4", ir.Yield)
	}
	if want := []string{"dinner", "quick"}; !reflect.DeepEqual(ir.Tags, want) {
		t.Errorf("Tags: want %v, got %v", want, ir.Tags)
	}
	wantDirections := []formats.TitledList{{Title: "", List: []string{"A directions"}}}
	if !reflect.DeepEqual(ir.Instructions, wantDirections) {
		t.Errorf("Instructions: want %v, got %v", wantDirections, ir.Instructions)
	}
	if len(ir.Images) != 1 {
		t.Errorf("Images: want 1, got %d", len(ir.Images))
	}
	if ir.PrepTime == nil || *ir.PrepTime != 20*time.Minute {
		t.Errorf("PrepTime: want 20m, got %v", ir.PrepTime)
	}
}

// TestExportSections covers the non-trivial part of Export: splitting a
// sectioned sequence back into ordered, cleanly-titled lists.
func TestExportSections(t *testing.T) {
	r := &paprika.Recipe{
		Directions: formats.SectionedSequence("Prep the base\n# Sauce\nMelt butter\nWhisk in flour"),
	}

	ir, err := r.Export()
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	want := []formats.TitledList{
		{Title: "", List: []string{"Prep the base"}},
		{Title: "Sauce", List: []string{"Melt butter", "Whisk in flour"}},
	}
	if !reflect.DeepEqual(ir.Instructions, want) {
		t.Errorf("Instructions: want %#v, got %#v", want, ir.Instructions)
	}
}

func TestStandardizeBookFromNotes(t *testing.T) {
	tests := []struct {
		name      string
		notes     string
		wantNotes string
		wantBook  utils.BookRef
	}{
		{"ISBN only", "ISBN: 9782019453411", "_9782019453411_",
			utils.BookRef{Book: utils.Book{ISBN13: "9782019453411"}}},
		{"ISBN, pages, recipe", "ISBN 978-3-16-148410-0\nPages 52\nRecipe 2", "_9783161484100, p.52, 2nd_",
			utils.BookRef{Book: utils.Book{ISBN13: "9783161484100"}, Pages: utils.Pages{utils.PageRange{"52"}}, RecipeNumber: 2}},
		{"No book reference", "Just a normal note.", "Just a normal note.", utils.BookRef{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &paprika.Recipe{Title: "X", Notes: tc.notes}
			if _, err := r.Standardize(); err != nil {
				t.Fatalf("standardize: %v", err)
			}
			if r.Notes != tc.wantNotes {
				t.Errorf("Notes: want %q, got %q", tc.wantNotes, r.Notes)
			}
			if got := r.Book(); !reflect.DeepEqual(got, tc.wantBook) {
				t.Errorf("Book: want %#v, got %#v", tc.wantBook, got)
			}
		})
	}
}

// TestParseFetchesRemoteImage covers the network-gated image_url fetch: a recipe
// with only a remote image keeps it unfetched unless network access is allowed.
func TestParseFetchesRemoteImage(t *testing.T) {
	pngBytes := onePixelPNG(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngBytes)
	}))
	defer srv.Close()

	body := `{"uid":"b","name":"Remote","description":"d","ingredients":"i",` +
		`"directions":"x","image_url":"` + srv.URL + `/photo.png"}`
	path := writeRecipeFile(t, t.TempDir(), "b.paprikarecipe", body)

	parseOne := func(opts formats.ParseOptions) *paprika.Recipe {
		t.Helper()
		pe, _, err := paprika.Parse(formats.Bundle{path}, opts)
		if err != nil {
			t.Fatal(err)
		}
		recipes := drain(t, pe)
		if len(recipes) != 1 {
			t.Fatalf("expected 1 recipe, got %d", len(recipes))
		}
		return recipes[0]
	}

	if off := parseOne(formats.ParseOptions{}); len(off.PhotoData) != 0 {
		t.Errorf("network off: PhotoData populated (%d bytes), want empty", len(off.PhotoData))
	}
	if on := parseOne(formats.ParseOptions{AllowNetwork: true}); len(on.PhotoData) == 0 {
		t.Error("network on: PhotoData empty, want the fetched image")
	}
}
