package mela

import (
	"strings"
	"time"
)

func durationGuesser(in string) (*time.Duration, error) {
	if in == "" {
		return nil, nil
	}

	in = strings.ReplaceAll(in, "hours", "h")
	in = strings.ReplaceAll(in, "hour", "h")
	in = strings.ReplaceAll(in, "mins", "m")
	in = strings.ReplaceAll(in, "min", "m")
	in = strings.ReplaceAll(in, ".", "")
	in = strings.ReplaceAll(in, " ", "")

	d, err := time.ParseDuration(in)
	if err != nil {
		return nil, err
	}

	return &d, nil
}
