package mela

import (
	"testing"
)

func Test_isbn10CheckDigit(t *testing.T) {
	type test struct {
		input string
		want  byte
	}

	tests := []test{
		{"JPJPJPJPJ", 0x0},
		{"019852666", '0'},
		{"019852660", '1'},
		{"019852665", '2'},
		{"019852656", '3'},
		{"019852664", '4'},
		{"019852669", '5'},
		{"019852663", '6'},
		{"019852668", '7'},
		{"019852662", '8'},
		{"019852667", '9'},
		{"019852661", 'X'},
	}
	for _, test := range tests {
		got := isbn10CheckDigit(test.input)

		if test.want != got {
			t.Errorf("Incorrect check digit for %s: want = %s, got = %s", test.input, string(test.want), string(got))
		}
	}
}

func Test_isbn13CheckDigit(t *testing.T) {
	type test struct {
		input string
		want  byte
	}

	tests := []test{
		{"JPJPJPJPJPJP", 0x0},
		{"978019852665", '0'},
		{"978019852668", '1'},
		{"978019852661", '2'},
		{"978019852664", '3'},
		{"978019852667", '4'},
		{"978019852660", '5'},
		{"978019852663", '6'},
		{"978019852666", '7'},
		{"978019852669", '8'},
		{"978019852662", '9'},
	}
	for _, test := range tests {
		got := isbn13CheckDigit(test.input)

		if test.want != got {
			t.Errorf("Incorrect check digit for %s: want = %s, got = %s", test.input, string(test.want), string(got))
		}
	}
}

func Test_validateISBN(t *testing.T) {
	type test struct {
		input    string
		wantISBN string
		wantErr  error
	}

	tests := []test{
		// Not ISBNs
		{"1234", "", ErrInvalidISBN},
		{"JPJP", "", ErrInvalidISBN},
		{"JPJPJPJPJP", "", ErrInvalidISBN10},
		{"JPJPJPJPJPJPJ", "", ErrInvalidISBN13},

		// Valid ISBN-10 -> ISBN-13
		{"0198526636", "9780198526636", nil},
		{"0 19 852663 6", "9780198526636", nil},
		{"0-19-852663-6", "9780198526636", nil},
		{"019852661X", "9780198526612", nil},
		{"019852661x", "9780198526612", nil},

		// Invalid ISBN-10
		{"0198526637", "", ErrIncorrectISBN10},

		// Invalid ISBN-13
		{"9780198526613", "", ErrIncorrectISBN13},
	}
	for _, test := range tests {
		gotISBN, gotErr := validateISBN(test.input)
		if gotErr != test.wantErr {
			t.Errorf("Error response incorrect for %s: want = %v, got = %v", test.input, test.wantErr, gotErr)
			continue
		}

		if test.wantISBN != gotISBN {
			t.Errorf("Incorrect ISBN for %s: want = %v, got = %v", test.input, test.wantISBN, gotISBN)
		}
	}
}
