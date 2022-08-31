package mela_test

import (
	"reflect"
	"testing"

	"github.com/jphastings/mela-recipes"
)

//

func TestRawRecipe_Book(t *testing.T) {
	type test struct {
		id   string
		want *mela.Book
	}

	tests := []test{
		{"urn:isbn:9782019453411#page=52", &mela.Book{ISBN13: "9782019453411", Page: 52}},
		{"urn:isbn:978-3-16-148410-0#page=1", nil},
		{"urn:isbn:malformed#page=52", nil},
		{"ACB628F3-DE6B-4833-A799-2B4F88CB0C1A", nil},
		{"example.org/path/to/something", nil},
	}

	for _, test := range tests {
		got := mela.RawRecipe{RawID: test.id}.Book()

		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("Incorrect book details for %s: want = %#v, got = %#v", test.id, test.want, got)
		}
	}
}

func TestRawRecipe_SetBook(t *testing.T) {
	type test struct {
		isbn    string
		page    uint
		wantID  string
		wantErr bool
	}

	tests := []test{
		{"9782019453411", 52, "urn:isbn:9782019453411#page=52", false},
		{"978-0-545-01022-1", 1024, "urn:isbn:9780545010221#page=1024", false},
		{"201945341X", 3, "urn:isbn:9782019453411#page=3", false},

		{"9782019453412", 1, "", true},
		{"2019453410", 1, "", true},
		{"notanisbn", 1, "", true},
	}

	for _, test := range tests {
		got, err := mela.RawRecipe{}.SetBook(test.isbn, test.page)

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

		if got.ID() != test.wantID {
			t.Errorf("Incorrect book ID for %s: want = %s, got = %s", test.isbn, test.wantID, got.ID())
		}
	}
}
