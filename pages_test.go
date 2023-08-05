package mela_test

import (
	"reflect"
	"testing"

	. "github.com/jphastings/mela-recipes"
)

func Test_ParsePages(t *testing.T) {
	type test struct {
		name string
		str  string
		want Pages
	}

	tests := []test{
		{"On single page", "52", Pages{PageRange{"52"}}},
		{"With a page range", "52,54", Pages{PageRange{"52"}, PageRange{"54"}}},
		{"With single pages that contain hyphens", "3%2D2", Pages{PageRange{"3-2"}}},
		{"With a page range", "52-56", Pages{PageRange{"52", "56"}}},
		{"With multiple pages", "42,52-56", Pages{PageRange{"42"}, PageRange{"52", "56"}}},
		{"With precent encoded", "3%2D2-3%2D4", Pages{PageRange{"3-2", "3-4"}}},

		{"With malformed page range", "52-56-60", nil},
		{"With malformed pages set", "52,,53", nil},
		{"With empty string", "", nil},
	}

	for _, test := range tests {
		got, err := ParsePages(test.str)
		if (err == nil) == (test.want == nil) {
			t.Errorf("Incorrect error state for %s: wantErr = %#v, got = %#v", test.name, test.want == nil, err)
		}

		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("Incorrect pages returned for %s: want = %#v, got = %#v", test.name, test.want, got)
		}
	}
}
