package mela_test

import (
	"reflect"
	"testing"

	"github.com/jphastings/recipes/internal/standardize"
	. "github.com/jphastings/recipes/mela"
	. "github.com/jphastings/recipes/utils"
	"github.com/stretchr/testify/assert"
)

func TestRawRecipe_Standardize(t *testing.T) {
	type test struct {
		name        string
		notes       string
		wantNotes   string
		wantBookRef BookRef
	}

	tests := []test{
		{"Just ISBN", "ISBN: 9782019453411", "_9782019453411_", BookRef{Book: Book{ISBN13: "9782019453411"}}},
		{"Just ISBN; simple", "_9782019453411_", "_9782019453411_", BookRef{Book: Book{ISBN13: "9782019453411"}}},
		{"ISBN and pages", "isbn: 978-3-16-148410-0\npages: 52", "_9783161484100, p.52_", BookRef{Book: Book{ISBN13: "9783161484100"}, Pages: Pages{PageRange{"52"}}}},
		{"ISBN and pages; simple", "_9783161484100, p.52_", "_9783161484100, p.52_", BookRef{Book: Book{ISBN13: "9783161484100"}, Pages: Pages{PageRange{"52"}}}},
		{"ISBN and multiple pages; simple", "_9783161484100, p.52-54_", "_9783161484100, p.52-54_", BookRef{Book: Book{ISBN13: "9783161484100"}, Pages: Pages{PageRange{"52", "54"}}}},
		{"ISBN, pages, recipe", "ISBN 978-3-16-148410-0\nPages 52\nRecipe 2", "_9783161484100, p.52, 2nd_", BookRef{Book: Book{ISBN13: "9783161484100"}, Pages: Pages{PageRange{"52"}}, RecipeNumber: 2}},
		{"ISBN, pages, recipe; simple", "_9783161484100, p.52, 2nd_", "_9783161484100, p.52, 2nd_", BookRef{Book: Book{ISBN13: "9783161484100"}, Pages: Pages{PageRange{"52"}}, RecipeNumber: 2}},
		{"Recipe, no pages", "ISBN: 978-3-16-148410-0\nRecipe: 2", "Recipe: 2\n\n_9783161484100_", BookRef{Book: Book{ISBN13: "9783161484100"}}},

		{"Text before", "Some other note.\n\nISBN: 9782019453411", "Some other note.\n\n_9782019453411_", BookRef{Book: Book{ISBN13: "9782019453411"}}},
		{"Text after", "ISBN: 9782019453411\n\nSome other note.", "Some other note.\n\n_9782019453411_", BookRef{Book: Book{ISBN13: "9782019453411"}}},
		{"Text both sides", "Something before.\n\nISBN: 9782019453411\n\n\nSomething after.", "Something before.\n\nSomething after.\n\n_9782019453411_", BookRef{Book: Book{ISBN13: "9782019453411"}}},

		{"Used in fixture", "C Notes\nISBN: 0198526636\npage 42\nrecipe: 3", "C Notes\n\n_9780198526636, p.42, 3rd_", BookRef{Book: Book{ISBN13: "9780198526636"}, Pages: Pages{PageRange{"42"}}, RecipeNumber: 3}},

		{"No details", "Some note mentioning an ISBN and pages and recipe.", "Some note mentioning an ISBN and pages and recipe.", BookRef{}},
	}

	fallbackBook := BookRef{Book: Book{ISBN13: "9781786699503"}}

	expStds := []standardize.Std{standardize.StdISBN}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := &Recipe{ID: "urn:isbn:" + fallbackBook.ISBN13, Notes: test.notes}

			stds, err := r.Standardize()
			assert.NoError(t, err)

			assert.Equal(t, expStds, stds)

			if reflect.DeepEqual(test.wantBookRef, Book{}) {
				test.wantBookRef = fallbackBook
			}

			assert.Equal(t, test.wantBookRef, r.Book())
			assert.Equal(t, test.wantNotes, r.Notes)
		})
	}
}
