package mela

import (
	"reflect"
	"testing"
)

//

func TestRawRecipe_Book(t *testing.T) {
	type test struct {
		name string
		id   string
		want *Book
	}

	tests := []test{
		{"With no page reference", "urn:isbn:9782019453411", &Book{ISBN13: "9782019453411", Pages: nil, RecipeNumber: 0}},
		{"With hyphenated ISBN", "urn:isbn:978-3-16-148410-0", &Book{ISBN13: "9783161484100", Pages: nil, RecipeNumber: 0}},
		{"With pages", "urn:isbn:9782019453411#pages=vii,1,4-8,3%2D2,3%2D4-3%2D6", &Book{
			ISBN13: "9782019453411",
			Pages: Pages{
				PageRange{"vii"},
				PageRange{"1"},
				PageRange{"4", "8"},
				PageRange{"3-2"},
				PageRange{"3-4", "3-6"},
			},
			RecipeNumber: 0}},
		{"With a recipe number", "urn:isbn:9782019453411#pages=52&recipe=2", &Book{ISBN13: "9782019453411", Pages: Pages{PageRange{"52"}}, RecipeNumber: 2}},
		{"With recipe number but no pages", "urn:isbn:9782019453411#recipe=2", &Book{ISBN13: "9782019453411"}},

		{"With invalid ISBN check digit", "urn:isbn:9782019453413", nil},
		{"With malformed ISBN", "urn:isbn:malformed#pages=52", nil},
		{"No ISBN", "ACB628F3-DE6B-4833-A799-2B4F88CB0C1A", nil},
		{"With URL", "example.org/path/to/something", nil},
	}

	for _, test := range tests {
		got := (&Recipe{ID: test.id}).Book()

		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("Incorrect book details for '%s': want = %#v, got = %#v", test.name, test.want, got)
		}
	}
}

func TestRawRecipe_SetBook(t *testing.T) {
	type test struct {
		name    string
		isbn    string
		pages   Pages
		index   uint
		wantID  string
		wantErr bool
	}

	tests := []test{
		{"With a page number", "9782019453411", Pages{PageRange{"42"}}, 0, "urn:isbn:9782019453411#pages=42", false},
		{"Over multiple pages", "978-0-545-01022-1", Pages{PageRange{"42"}, PageRange{"52", "56"}}, 1, "urn:isbn:9780545010221#pages=42,52-56&recipe=1", false},
		{"With an ISBN10", "201945341X", nil, 0, "urn:isbn:9782019453411", false},

		{"With an invalid ISBN check digit", "9782019453412", nil, 0, "", true},
		{"With an invalid ISBN", "2019453410", nil, 0, "", true},
		{"With a non-ISBN", "notanisbn", nil, 0, "", true},
	}

	for _, test := range tests {

		r := &Recipe{}
		err := r.SetBook(test.isbn, test.pages, test.index)

		if test.wantErr {
			if err == nil {
				t.Errorf("Unexpectedly got no error for %s", test.isbn)
			}
			continue
		}

		if !test.wantErr && err != nil {
			t.Errorf("Got unexpected error for %s: got = %v", test.isbn, err)
			continue
		}

		if r.ID != test.wantID {
			t.Errorf("Incorrect book ID for %s: want = %s, got = %s", test.isbn, test.wantID, r.ID)
		}
	}
}
