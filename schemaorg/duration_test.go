package schemaorg

import (
	"testing"
	"time"
)

func TestParseISODuration(t *testing.T) {
	dur := func(d time.Duration) *time.Duration { return &d }

	cases := []struct {
		in   string
		want *time.Duration
	}{
		{"PT1H30M", dur(90 * time.Minute)},
		{"PT30M", dur(30 * time.Minute)},
		{"PT45S", dur(45 * time.Second)},
		{"P0DT0H20M", dur(20 * time.Minute)},
		{"P1DT2H", dur(26 * time.Hour)},
		{"PT2H30M15S", dur(2*time.Hour + 30*time.Minute + 15*time.Second)},
		{"  PT15M  ", dur(15 * time.Minute)},
		{"", nil},
		{"30 min", nil},  // not ISO-8601
		{"PT0S", nil},    // zero → nil
		{"P", nil},       // empty duration
		{"P1Y", nil},     // years not accepted
		{"garbage", nil}, // unparseable
	}

	for _, c := range cases {
		got := parseISODuration(c.in)
		switch {
		case c.want == nil && got != nil:
			t.Errorf("parseISODuration(%q) = %v, want nil", c.in, *got)
		case c.want != nil && got == nil:
			t.Errorf("parseISODuration(%q) = nil, want %v", c.in, *c.want)
		case c.want != nil && got != nil && *got != *c.want:
			t.Errorf("parseISODuration(%q) = %v, want %v", c.in, *got, *c.want)
		}
	}
}
