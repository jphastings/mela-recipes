package mela

import (
	"regexp"
	"strings"
)

type SectionedSequence string

var sectionDivider = regexp.MustCompile(`#+\s+(.+)`)

func (ss SectionedSequence) Parse() map[string][]string {
	sections := make(map[string][]string)
	heading := ""
	for _, line := range strings.Split(string(ss), "\n") {
		if newHeading := sectionDivider.FindString(line); newHeading != "" {
			heading = newHeading
			continue
		}

		sections[heading] = append(sections[heading], line)
	}
	return sections
}
