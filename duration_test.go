package mela

import (
	"testing"
	"time"
)

func Test_durationGuesser(t *testing.T) {
	type test struct {
		input   string
		want    *time.Duration
		wantErr bool
	}

	hour1 := time.Hour
	hour2 := 2 * time.Hour
	min1 := time.Minute
	min2 := 2 * time.Minute

	tests := []test{
		{"", nil, false},

		{"1h", &hour1, false},
		{"2h", &hour2, false},
		{"1 hour", &hour1, false},
		{"1 hours", &hour1, false},
		{"2 hour", &hour2, false},
		{"2 hours", &hour2, false},
		{"1hour", &hour1, false},
		{"1h.", &hour1, false},

		{"1m", &min1, false},
		{"2m", &min2, false},
		{"1 min", &min1, false},
		{"1 min.", &min1, false},
		{"1 mins", &min1, false},
		{"1 mins.", &min1, false},
		{"2 min", &min2, false},
		{"2 min.", &min2, false},
		{"2 mins", &min2, false},
		{"2 mins.", &min2, false},

		{"nope", nil, true},
	}
	for _, test := range tests {
		got, err := durationGuesser(test.input)
		if err != nil {
			if !test.wantErr {
				t.Errorf("Expected no error, got = %v", err)
			}
			continue
		}

		if (test.want == nil && got != nil) || (test.want != nil && *got != *test.want) {
			t.Errorf("Incorrect output: want = %v, got = %v", test.want, got)
		}
	}
}
