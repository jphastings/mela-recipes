package epub

import "testing"

func TestTidyText(t *testing.T) {
	cases := map[string]string{
		"a\n\n\nb":        "a\n\nb",      // 3 newlines -> blank line
		"a\n \n\nb":       "a\n\nb",      // whitespace between newlines
		"a\n\n\n\nc":      "a\n\nc",      // 4 newlines -> blank line
		"\n\nStart\n\n":   "Start",       // leading/trailing trimmed
		"  \n\ntitle":     "title",       // leading whitespace + blank line
		"one\ntwo":        "one\ntwo",    // single newline preserved
		"sub\n\ndesc":     "sub\n\ndesc", // already a single blank line
		"sub\n \t \ndesc": "sub\n\ndesc", // blank line of only whitespace
	}
	for in, want := range cases {
		if got := tidyText(in); got != want {
			t.Errorf("tidyText(%q) = %q, want %q", in, got, want)
		}
	}
}
