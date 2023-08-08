package mela

import (
	"strings"
	"time"
)

type MaybeDuration string

func (m MaybeDuration) Parse() (*time.Duration, error) {
	if m == "" {
		return nil, nil
	}

	in := strings.ReplaceAll(string(m), "hours", "h")
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
