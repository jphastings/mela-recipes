package paprika

import (
	"errors"

	"github.com/jphastings/recipes/internal/standardize"
	"github.com/jphastings/recipes/utils"
)

// Standardize applies every standardization in a best-effort, fail-soft manner:
// a failure in one step is recorded but does not prevent the others from
// running. The returned slice lists only the standardizations that were actually
// applied; the returned error reports any steps that failed.
func (r *Recipe) Standardize() ([]standardize.Std, error) {
	var stds []standardize.Std
	var errs []error

	var stdApplied bool
	if r.filename, stdApplied = standardize.Filename(r.filename, r.Title); stdApplied {
		stds = append(stds, standardize.StdFilename)
	}

	newNotes, book, found, err := utils.ExtractBookFromNotes(r.Notes)
	if err != nil {
		errs = append(errs, err)
	} else if found {
		r.Notes = newNotes
		r.UID = book.URN()
		stds = append(stds, standardize.StdISBN)
	}

	if r.Categories == nil {
		r.Categories = make([]string, 0)
	}

	if len(r.PhotoData) > 0 {
		images := []utils.B64Image{r.PhotoData}
		resized, err := utils.OptimizeImages(images)
		if err != nil {
			errs = append(errs, err)
		} else {
			r.PhotoData = images[0]
		}
		if resized {
			stds = append(stds, standardize.StdImages)
		}
	}

	return stds, errors.Join(errs...)
}
