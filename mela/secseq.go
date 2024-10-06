package mela

import (
	"regexp"
	"strings"
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

func NewSectionedSequence(items map[string][]string) SectionedSequence {
	var ss []string

	if lines, ok := items[""]; ok {
		ss = append(ss, lines...)
	}
	for heading, lines := range items {
		if heading == "" {
			continue
		}

		ss = append(ss, "# "+heading)
		ss = append(ss, lines...)
	}
	return SectionedSequence(strings.Join(ss, "\n"))
}
