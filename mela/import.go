package mela

import (
	"fmt"
	"regexp"

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
		filename:     formats.WithoutExt(r.Filename()),
		ID:           id,
		Title:        ir.Title,
		Link:         ir.Source.URI,
		Text:         melaText(ir.Description),
		Ingredients:  formats.IngredientsToSectioned(ir.Ingredients),
		Instructions: formats.TitledListsToSectioned(ir.Instructions),
		Notes:        melaNotes(ir),
		Images:       ir.Images,

		Categories: importTags(ir.Tags),
		Yield:      PeopleCount(ir.Yield),

		PrepTime:  formats.FormatDuration(ir.PrepTime),
		CookTime:  formats.FormatDuration(ir.CookTime),
		TotalTime: formats.FormatDuration(ir.TotalTime),
	}

	return mr, nil
}

// blankLine matches a run of two or more newlines with any interspersed
// horizontal whitespace.
var blankLine = regexp.MustCompile(`[^\S\n]*\n[^\S\n]*(?:\n[^\S\n]*)+`)

// melaText prepares a description for Mela, which renders newlines literally: a
// paragraph break ("\n\n", the interchange/Markdown convention) would show as an
// empty line, so collapse every run of two or more newlines to a single one.
func melaText(s string) string {
	return blankLine.ReplaceAllString(s, "\n")
}

// melaNotes carries the interchange notes and, since Mela has no dedicated
// source field, appends the source's name as an attribution line (eg. the book a
// recipe was extracted from). The source URI itself becomes the Mela link.
func melaNotes(ir formats.InterchangeRecipe) string {
	notes := ir.Notes
	if ir.Source.Name != "" {
		if notes != "" {
			notes += "\n\n"
		}
		notes += "From " + ir.Source.Name
	}
	return notes
}

// importTags carries the interchange recipe's tags across as Mela categories,
// always returning a non-nil slice.
func importTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}
