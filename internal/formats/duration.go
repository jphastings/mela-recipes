package formats

import (
	"fmt"
	"strings"
	"time"
)

type MaybeDuration string

// FormatDuration renders an optional duration as a MaybeDuration in "<h>h<m>m"
// form (an absent duration becomes the empty value). The minute component is
// kept within the hour so the result round-trips back through Parse.
func FormatDuration(dur *time.Duration) MaybeDuration {
	if dur == nil {
		return ""
	}

	hours := int(dur.Hours())
	mins := int(dur.Minutes()) % 60
	return MaybeDuration(fmt.Sprintf("%dh%dm", hours, mins))
}

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
