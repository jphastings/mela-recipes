package mela_test

import (
	"reflect"
	"testing"

	. "github.com/jphastings/mela-recipes"
)

func TestRawRecipe_Standardize(t *testing.T) {
	type test struct {
		name      string
		notes     string
		wantNotes string
		wantBook  *Book
	}

	tests := []test{
		{"Just ISBN", "ISBN: 9782019453411", "_9782019453411_", &Book{ISBN13: "9782019453411"}},
		{"Just ISBN; simple", "_9782019453411_", "_9782019453411_", &Book{ISBN13: "9782019453411"}},
		{"ISBN and pages", "isbn: 978-3-16-148410-0\npages: 52", "_9783161484100, p.52_", &Book{ISBN13: "9783161484100", Pages: Pages{PageRange{"52"}}}},
		{"ISBN and pages; simple", "_9783161484100, p.52_", "_9783161484100, p.52_", &Book{ISBN13: "9783161484100", Pages: Pages{PageRange{"52"}}}},
		{"ISBN, pages, recipe", "ISBN 978-3-16-148410-0\nPages 52\nRecipe 2", "_9783161484100, p.52, 2nd_", &Book{ISBN13: "9783161484100", Pages: Pages{PageRange{"52"}}, RecipeNumber: 2}},
		{"ISBN, pages, recipe; simple", "_9783161484100, p.52, 2nd_", "_9783161484100, p.52, 2nd_", &Book{ISBN13: "9783161484100", Pages: Pages{PageRange{"52"}}, RecipeNumber: 2}},
		{"Recipe, no pages", "ISBN: 978-3-16-148410-0\nRecipe: 2", "Recipe: 2\n\n_9783161484100_", &Book{ISBN13: "9783161484100"}},

		{"Text before", "Some other note.\n\nISBN: 9782019453411", "Some other note.\n\n_9782019453411_", &Book{ISBN13: "9782019453411"}},
		{"Text after", "ISBN: 9782019453411\n\nSome other note.", "Some other note.\n\n_9782019453411_", &Book{ISBN13: "9782019453411"}},
		{"Text both sides", "Something before.\n\nISBN: 9782019453411\n\n\nSomething after.", "Something before.\n\nSomething after.\n\n_9782019453411_", &Book{ISBN13: "9782019453411"}},

		{"Used in fixture", "C Notes\nISBN: 0198526636\npage 42\nrecipe: 3", "C Notes\n\n_9780198526636, p.42, 3rd_", &Book{ISBN13: "9780198526636", Pages: Pages{PageRange{"42"}}, RecipeNumber: 3}},

		{"No details", "Some note mentioning an ISBN and pages and recipe.", "Some note mentioning an ISBN and pages and recipe.", nil},
	}

	fallbackBook := &Book{ISBN13: "9781786699503"}

	for _, test := range tests {
		r := &Recipe{ID: "urn:isbn:9781786699503", Notes: test.notes}
		if err := r.Standardize(); err != nil {
			t.Errorf("Error standardizing for '%s': %v", test.name, err)
		}

		if test.wantBook == nil {
			test.wantBook = fallbackBook
		}

		if !reflect.DeepEqual(test.wantBook, r.Book()) {
			t.Errorf("Incorrect book details for '%s': want = %#v, got = %#v", test.name, test.wantBook, r.Book())
		}
		if r.Notes != test.wantNotes {
			t.Errorf("Incorrect book notes for '%s': want = %#v, got = %#v", test.name, test.wantNotes, r.Notes)
		}
	}

}
