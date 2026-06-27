package schemaorg

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// isoDuration matches the subset of ISO-8601 durations that recipe times use:
// weeks, days, hours, minutes and (fractional) seconds. Years and months are
// ambiguous in absolute length and aren't meaningful for cooking times, so they
// are not accepted.
var isoDuration = regexp.MustCompile(`^P(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`)

// parseISODuration parses an ISO-8601 duration such as "PT1H30M" into a
// *time.Duration. It returns nil for an empty, zero, or unparseable value, since
// the interchange format treats a missing duration as a nil pointer.
func parseISODuration(s string) *time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	m := isoDuration.FindStringSubmatch(s)
	if m == nil {
		return nil
	}

	var d time.Duration
	add := func(group string, unit time.Duration) {
		if group == "" {
			return
		}
		if n, err := strconv.Atoi(group); err == nil {
			d += time.Duration(n) * unit
		}
	}
	add(m[1], 7*24*time.Hour)
	add(m[2], 24*time.Hour)
	add(m[3], time.Hour)
	add(m[4], time.Minute)
	if m[5] != "" {
		if sec, err := strconv.ParseFloat(m[5], 64); err == nil {
			d += time.Duration(sec * float64(time.Second))
		}
	}

	if d == 0 {
		return nil
	}
	return &d
}
