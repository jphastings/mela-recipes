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
		{"Just ISBN", "ISBN: 9782019453411", "", &Book{ISBN13: "9782019453411"}},
		{"ISBN and pages", "isbn: 978-3-16-148410-0\npages: 52", "", &Book{ISBN13: "9783161484100", Pages: Pages{PageRange{"52"}}}},
		{"Page", "isbn 978-3-16-148410-0\npage 52", "", &Book{ISBN13: "9783161484100", Pages: Pages{PageRange{"52"}}}},
		{"ISBN, pages, recipe", "ISBN 978-3-16-148410-0\nPages 52\nRecipe 2", "", &Book{ISBN13: "9783161484100", Pages: Pages{PageRange{"52"}}, RecipeNumber: 2}},
		{"Recipe, no pages", "ISBN: 978-3-16-148410-0\nRecipe: 2", "Recipe: 2", &Book{ISBN13: "9783161484100"}},

		{"Text before", "Some other note.\n\nISBN: 9782019453411", "Some other note.", &Book{ISBN13: "9782019453411"}},
		{"Text after", "ISBN: 9782019453411\n\nSome other note.", "Some other note.", &Book{ISBN13: "9782019453411"}},
		{"Text both sides", "Something before.\n\nISBN: 9782019453411\n\n\nSomething after.", "Something before.\n\nSomething after.", &Book{ISBN13: "9782019453411"}},

		{"Used in fixture", "C Notes\nISBN: 0198526636\npage 42\nrecipe: 3", "C Notes", &Book{ISBN13: "9780198526636", Pages: Pages{PageRange{"42"}}, RecipeNumber: 3}},

		{"No details", "Some note mentioning an ISBN and pages and recipe.", "Some note mentioning an ISBN and pages and recipe.", nil},
	}

	for _, test := range tests {
		r := (&RawRecipe{RawNotes: test.notes}).Standardize()

		if !reflect.DeepEqual(test.wantBook, r.Book()) {
			t.Errorf("Incorrect book details for '%s': want = %#v, got = %#v", test.name, test.wantBook, r.Book())
		}
		if r.Notes() != test.wantNotes {
			t.Errorf("Incorrect book notes for '%s': want = %#v, got = %#v", test.name, test.wantNotes, r.Notes())
		}
	}

	r := (&RawRecipe{
		RawID:    "urn:isbn:9782019453411",
		RawNotes: "Some note.\n\nISBN: 9783161484100",
	}).Standardize()

	wantISBN := "9783161484100"
	if r.Book().ISBN13 != wantISBN {
		t.Errorf("Book details in notes incorrectly ignored (when ID is already ISBN): want = %#v, got = %#v", wantISBN, r.Book().ISBN13)
	}
}
