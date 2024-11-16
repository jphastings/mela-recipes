package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//

func TestNewBookRefFromURN(t *testing.T) {
	type test struct {
		name    string
		urn     string
		want    BookRef
		wantErr bool
	}

	tests := []test{
		{"With no page reference", "urn:isbn:9782019453411", BookRef{Book: Book{ISBN13: "9782019453411"}, Pages: nil, RecipeNumber: 0}, false},
		{"With hyphenated ISBN", "urn:isbn:978-3-16-148410-0", BookRef{Book: Book{ISBN13: "9783161484100"}, Pages: nil, RecipeNumber: 0}, false},
		{"With pages", "urn:isbn:9782019453411#pages=vii,1,4-8,3%2D2,3%2D4-3%2D6", BookRef{
			Book: Book{ISBN13: "9782019453411"},
			Pages: Pages{
				PageRange{"vii"},
				PageRange{"1"},
				PageRange{"4", "8"},
				PageRange{"3-2"},
				PageRange{"3-4", "3-6"},
			},
			RecipeNumber: 0}, false},
		{"With a recipe number", "urn:isbn:9782019453411#pages=52&recipe=2", BookRef{Book: Book{ISBN13: "9782019453411"}, Pages: Pages{PageRange{"52"}}, RecipeNumber: 2}, false},
		{"With recipe number but no pages", "urn:isbn:9782019453411#recipe=2", BookRef{Book: Book{ISBN13: "9782019453411"}}, false},

		{"With invalid ISBN check digit", "urn:isbn:9782019453413", BookRef{}, true},
		{"With malformed ISBN", "urn:isbn:malformed#pages=52", BookRef{}, true},
		{"No ISBN", "ACB628F3-DE6B-4833-A799-2B4F88CB0C1A", BookRef{}, true},
		{"With URL", "example.org/path/to/something", BookRef{}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewBookRefFromURN(test.urn)

			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.want, got)
		})
	}
}

func TestBook_URN(t *testing.T) {
	type test struct {
		name    string
		ref     BookRef
		wantURN string
	}

	bookA := Book{ISBN13: "9782019453411"}
	bookB := Book{ISBN13: "9780545010221"}

	tests := []test{
		{"Only the ISBN", BookRef{Book: bookA}, "urn:isbn:9782019453411"},
		{"With a page number", BookRef{Book: bookA, Pages: Pages{PageRange{"42"}}}, "urn:isbn:9782019453411#pages=42"},
		{"Over multiple pages", BookRef{Book: bookB, Pages: Pages{PageRange{"42"}, PageRange{"52", "56"}}, RecipeNumber: 1}, "urn:isbn:9780545010221#pages=42,52-56&recipe=1"},

		{"recipe number requires pages", BookRef{Book: bookB, RecipeNumber: 1}, "urn:isbn:9780545010221"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urn := test.ref.URN()

			assert.Equal(t, test.wantURN, urn)
		})
	}
}
