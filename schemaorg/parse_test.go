package schemaorg

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jphastings/recipes/internal/formats"
)

func dur(d time.Duration) *time.Duration { return &d }

func load(t *testing.T, name string) formats.InterchangeRecipe {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("fixtures", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	ir, err := extractRecipe(data, strings.ToLower(filepath.Ext(name)))
	if err != nil {
		t.Fatalf("extractRecipe(%s): %v", name, err)
	}
	return ir
}

type section struct {
	title string
	steps int
}

func TestExtract(t *testing.T) {
	tests := []struct {
		file              string
		title             string
		desc              string
		yield             string
		ingredients       int
		sections          []section
		prep, cook, total *time.Duration
		tags              []string
		sourceName        string
		sourceURI         string
	}{
		{
			file:        "jsonld.html",
			title:       "Classic Pancakes",
			desc:        "Fluffy and quick pancakes.", // HTML tags stripped
			yield:       "12 pancakes",
			ingredients: 4,
			sections:    []section{{"", 3}}, // HowToStep array → one untitled section
			prep:        dur(10 * time.Minute),
			cook:        dur(15 * time.Minute),
			total:       dur(25 * time.Minute),
			tags:        []string{"Breakfast", "American", "pancakes", "easy"}, // deduped case-insensitively
			sourceName:  "Sam Cook",
			sourceURI:   "https://example.com/pancakes",
		},
		{
			file:        "howtosection.json",
			title:       "Layered Trifle",
			yield:       "8 servings",
			ingredients: 4,
			sections:    []section{{"Custard", 2}, {"Assembly", 2}},
			total:       dur(2 * time.Hour),
		},
		{
			file:        "nextcloud.json",
			title:       "Tomato Soup",
			desc:        "A simple weeknight soup.",
			yield:       "4", // number stringified
			ingredients: 4,
			sections:    []section{{"", 3}}, // array of strings → one untitled section
			prep:        dur(5 * time.Minute),
			cook:        dur(25 * time.Minute),
			tags:        []string{"Soup"},
			sourceURI:   "http://nextcloud.local/apps/cookbook/recipes/3",
		},
		{
			file:        "microdata.html",
			title:       "Guacamole",
			desc:        "Fresh and zesty.",
			yield:       "2 cups",
			ingredients: 3,
			sections:    []section{{"", 2}},
			prep:        dur(10 * time.Minute),
			tags:        []string{"Dip"},
			sourceName:  "Alex", // nested author itemscope, not the recipe name
		},
		{
			file:        "hrecipe.html",
			title:       "Lemonade",
			desc:        "A refreshing summer drink.",
			yield:       "1 jug",
			ingredients: 3,
			sections:    []section{{"", 3}}, // single e-instructions block split into steps
			total:       dur(5 * time.Minute),
			tags:        []string{"Drinks"},
			sourceName:  "Jo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			ir := load(t, tc.file)

			if ir.Title != tc.title {
				t.Errorf("Title = %q, want %q", ir.Title, tc.title)
			}
			if tc.desc != "" && ir.Description != tc.desc {
				t.Errorf("Description = %q, want %q", ir.Description, tc.desc)
			}
			if tc.yield != "" && ir.Yield != tc.yield {
				t.Errorf("Yield = %q, want %q", ir.Yield, tc.yield)
			}
			if n := countItems(ir.Ingredients); n != tc.ingredients {
				t.Errorf("ingredient count = %d, want %d", n, tc.ingredients)
			}

			if len(ir.Instructions) != len(tc.sections) {
				t.Fatalf("instruction sections = %d, want %d (%+v)", len(ir.Instructions), len(tc.sections), ir.Instructions)
			}
			for i, sec := range tc.sections {
				if ir.Instructions[i].Title != sec.title {
					t.Errorf("section %d title = %q, want %q", i, ir.Instructions[i].Title, sec.title)
				}
				if got := len(ir.Instructions[i].List); got != sec.steps {
					t.Errorf("section %d steps = %d, want %d", i, got, sec.steps)
				}
			}

			assertDur(t, "prepTime", ir.PrepTime, tc.prep)
			assertDur(t, "cookTime", ir.CookTime, tc.cook)
			assertDur(t, "totalTime", ir.TotalTime, tc.total)

			if tc.tags != nil && !equalStrings(ir.Tags, tc.tags) {
				t.Errorf("Tags = %v, want %v", ir.Tags, tc.tags)
			}
			if tc.sourceName != "" && ir.Source.Name != tc.sourceName {
				t.Errorf("Source.Name = %q, want %q", ir.Source.Name, tc.sourceName)
			}
			if tc.sourceURI != "" && ir.Source.URI != tc.sourceURI {
				t.Errorf("Source.URI = %q, want %q", ir.Source.URI, tc.sourceURI)
			}

			if errs := ir.Validate(); len(errs) > 0 {
				t.Errorf("Validate() returned errors: %v", errs)
			}
		})
	}
}

func TestParseEmitsRecipe(t *testing.T) {
	pe, cd, err := Parse(formats.Bundle{filepath.Join("fixtures", "jsonld.html")}, formats.ParseOptions{})
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
	if got == nil {
		t.Fatal("no recipe emitted")
	}
	if got.Name() != "Classic Pancakes" {
		t.Errorf("Name() = %q, want Classic Pancakes", got.Name())
	}
	if _, err := got.Standardize(); err != nil {
		t.Errorf("Standardize: %v", err)
	}
	if fn := got.Filename(); !strings.HasSuffix(fn, ".html") || !strings.Contains(fn, "pancakes") {
		t.Errorf("Filename() = %q, want a .html name derived from the title", fn)
	}
}

func TestExtractNoRecipe(t *testing.T) {
	_, err := extractRecipe([]byte("<html><body><p>nothing structured here</p></body></html>"), ".html")
	if !errors.Is(err, errNoRecipe) {
		t.Errorf("got %v, want errNoRecipe", err)
	}
}

func countItems(lists []formats.TitledList) int {
	n := 0
	for _, l := range lists {
		n += len(l.List)
	}
	return n
}

func assertDur(t *testing.T, name string, got, want *time.Duration) {
	t.Helper()
	switch {
	case want == nil && got != nil:
		t.Errorf("%s = %v, want nil", name, *got)
	case want != nil && got == nil:
		t.Errorf("%s = nil, want %v", name, *want)
	case want != nil && got != nil && *got != *want:
		t.Errorf("%s = %v, want %v", name, *got, *want)
	}
}

func equalStrings(a, b []string) bool {
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
