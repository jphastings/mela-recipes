package mela

import (
	"errors"
	"sort"

	"github.com/jphastings/recipes/internal/formats"
)

// Export converts a Mela recipe into the interchange format. It is the inverse
// of ImportRecipe; fields the interchange format cannot represent (eg. link,
// nutrition) are dropped. The interchange filename is left unset, as that field
// has no exported setter.
func (r *Recipe) Export() (formats.InterchangeRecipe, error) {
	ir := formats.NewInterchangeRecipe()
	ir.ID = r.ID
	ir.Title = r.Title
	ir.Description = r.Text
	ir.Notes = r.Notes
	ir.Yield = string(r.Yield)
	ir.Ingredients = secSeqToTitledLists(r.Ingredients)
	ir.Instructions = secSeqToTitledLists(r.Instructions)

	if r.Categories != nil {
		ir.Tags = r.Categories
	}
	if r.Images != nil {
		ir.Images = r.Images
	}

	prep, prepErr := r.PrepTime.Parse()
	cook, cookErr := r.CookTime.Parse()
	total, totalErr := r.TotalTime.Parse()
	ir.PrepTime = prep
	ir.CookTime = cook
	ir.TotalTime = total

	return ir, errors.Join(prepErr, cookErr, totalErr)
}

// secSeqToTitledLists is the inverse of titledListsToSecSeq: it splits a
// SectionedSequence back into titled lists, with the untitled section first and
// any named sections following in a stable (sorted) order.
func secSeqToTitledLists(ss SectionedSequence) []formats.TitledList {
	sections := ss.Parse()

	var lists []formats.TitledList
	if lines, ok := sections[""]; ok {
		lists = append(lists, formats.TitledList{Title: "", List: lines})
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
		lists = append(lists, formats.TitledList{Title: title, List: sections[heading]})
	}

	return lists
}
