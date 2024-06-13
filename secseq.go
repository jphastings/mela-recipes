package mela

import (
	"regexp"
)

type SectionedSequence string

var (
	sectionDivider = regexp.MustCompile(`#+\s+(.+)`)
	lineDivider    = regexp.MustCompile(`\n+`)
)

func (ss SectionedSequence) Parse() map[string][]string {
	sections := make(map[string][]string)
	heading := ""
	for _, line := range lineDivider.Split(string(ss), -1) {
		if newHeading := sectionDivider.FindString(line); newHeading != "" {
			heading = newHeading
			continue
		}

		sections[heading] = append(sections[heading], line)
	}
	return sections
}
