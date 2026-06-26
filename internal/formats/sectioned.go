package formats

import (
	"regexp"
	"sort"
	"strings"
)

// SectionedSequence is a newline-delimited list of items (ingredients or
// instructions) in which a line of the form "# Heading" starts a named section.
// Several recipe formats (Mela, Paprika) store their ingredients and directions
// this way, so the type and its converters live here for sharing.
type SectionedSequence string

var (
	sectionDivider = regexp.MustCompile(`#+\s+(.+)`)
	lineDivider    = regexp.MustCompile(`\n+`)
)

// Parse splits the sequence into its sections, keyed by the heading line (the
// untitled leading section uses the empty-string key).
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

// NewSectionedSequence builds a sequence from a section map; the empty-string
// key (if present) leads, followed by each named section under a "# Heading".
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

// TitledListsToSectioned flattens interchange titled lists into a sectioned
// sequence.
func TitledListsToSectioned(tls []TitledList) SectionedSequence {
	m := make(map[string][]string)
	for _, tl := range tls {
		m[tl.Title] = tl.List
	}
	return NewSectionedSequence(m)
}

// SectionedToTitledLists is the inverse of TitledListsToSectioned: it splits a
// sequence back into titled lists, with the untitled section first and any
// named sections following in a stable (sorted) order, their "# " markers
// stripped from the titles.
func SectionedToTitledLists(ss SectionedSequence) []TitledList {
	sections := ss.Parse()

	var lists []TitledList
	if lines, ok := sections[""]; ok {
		lists = append(lists, TitledList{Title: "", List: lines})
	}

	headings := make([]string, 0, len(sections))
	for heading := range sections {
		if heading != "" {
			headings = append(headings, heading)
		}
	}
	sort.Strings(headings)

	for _, heading := range headings {
		title := heading
		if m := sectionDivider.FindStringSubmatch(heading); m != nil {
			title = m[1]
		}
		lists = append(lists, TitledList{Title: title, List: sections[heading]})
	}

	return lists
}
