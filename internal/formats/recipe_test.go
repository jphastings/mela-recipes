package formats

import "testing"

func TestDisplayTitle(t *testing.T) {
	cases := []struct {
		title, subtitle, want string
	}{
		{"Abricotines", "Apricot Balls", "Abricotines (Apricot Balls)"},
		{"Saag Paneer", "Spinach with Indian cheese", "Saag Paneer (Spinach with Indian cheese)"},
		{"Pancakes", "", "Pancakes"},
	}
	for _, c := range cases {
		ir := InterchangeRecipe{Title: c.title, Subtitle: c.subtitle}
		if got := ir.DisplayTitle(); got != c.want {
			t.Errorf("DisplayTitle(%q, %q) = %q, want %q", c.title, c.subtitle, got, c.want)
		}
	}
}
