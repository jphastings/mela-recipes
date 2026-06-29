package ingredients

import "testing"

func TestNormalizeAmount(t *testing.T) {
	cases := map[string]string{
		"1/2":    "½",
		"3/4":    "¾",
		"1 1/2":  "1½",
		"2":      "2",
		"3.5":    "3.5",
		"1/16":   "0.0625",
		"2 1/16": "2.0625",
		"":       "",
	}
	for in, want := range cases {
		if got := NormalizeAmount(in); got != want {
			t.Errorf("NormalizeAmount(%q) = %q, want %q", in, got, want)
		}
	}
}
