package mela

import (
	"fmt"
	"time"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/uuid"
)

func ImportRecipe(r formats.Recipe) (formats.Recipe, error) {
	if r == nil {
		return nil, fmt.Errorf("provided recipe is nil")
	}

	// If its already in this format then no conversion is needed
	if _, ok := r.(*Recipe); ok {
		return r, nil
	}

	ir, err := r.Export()
	if err != nil {
		return nil, err
	}

	id := ir.ID
	if id == "" {
		if newID, err := uuid.NewUUID(ir.Title); err == nil {
			id = newID.String()
		}
	}

	mr := &Recipe{
		filename: r.Filename(),
		ID:       id,
		Title:    ir.Title,
		// TODO: Link (source)
		Text:         ir.Description,
		Ingredients:  titledListsToSecSeq(ir.Ingredients),
		Instructions: titledListsToSecSeq(ir.Instructions),
		Images:       ir.Images,

		Categories: []string{},
		Yield:      PeopleCount(ir.Yield),

		PrepTime:  formatDuration(ir.PrepTime),
		CookTime:  formatDuration(ir.CookTime),
		TotalTime: formatDuration(ir.TotalTime),
	}

	return mr, nil
}

func formatDuration(dur *time.Duration) formats.MaybeDuration {
	if dur == nil {
		return ""
	}

	hours := int(dur.Hours())
	mins := int(dur.Minutes())
	return formats.MaybeDuration(fmt.Sprintf("%dh%dm", hours, mins))
}

func titledListsToSecSeq(tls []formats.TitledList) SectionedSequence {
	m := make(map[string][]string)
	for _, tl := range tls {
		m[tl.Title] = tl.List
	}

	return NewSectionedSequence(m)
}
