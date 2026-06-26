package mela

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jphastings/recipes/internal/standardize"
	"github.com/jphastings/recipes/utils"
)

// Standardize applies every standardization in a best-effort, fail-soft manner:
// a failure in one step is recorded but does not prevent the others from running.
// The returned slice lists only the standardizations that were actually applied,
// and the returned error (via errors.Join) reports any steps that failed.
func (r *Recipe) Standardize() ([]standardize.Std, error) {
	var stds []standardize.Std
	var errs []error

	var stdApplied bool
	if r.filename, stdApplied = standardize.Filename(r.filename, r.Title); stdApplied {
		stds = append(stds, standardize.StdFilename)
	}

	if applied, err := bookFromNotes(r); err != nil {
		errs = append(errs, err)
	} else if applied {
		stds = append(stds, standardize.StdISBN)
	}

	if r.Images == nil {
		r.Images = make([]utils.B64Image, 0)
	}

	if r.Categories == nil {
		r.Categories = make([]string, 0)
	}

	var imagesResized bool
	for i, img := range r.Images {
		newImg, ok, err := img.Optimize()
		if err != nil {
			errs = append(errs, fmt.Errorf("image %d: %w", i, err))
			continue
		}
		r.Images[i] = newImg
		imagesResized = imagesResized || ok
	}
	if imagesResized {
		stds = append(stds, standardize.StdImages)
	}

	return stds, errors.Join(errs...)
}

var extractor = regexp.MustCompile(`(?i)(\s*)((?:isbn:? ?|_)([0-9X-]+)\r?\n?((?:, p\.|pages?:? ?)([^_\s,]+)\r?\n?((?:recipe:? ?|, )?(\d+)(?:[a-z]{2})?\r?\n?)?)?)_?(\s*)`)

func bookFromNotes(r *Recipe) (bool, error) {
	matches := extractor.FindStringSubmatch(r.Notes)
	if matches == nil {
		return false, nil
	}

	var newNotes string
	around := strings.SplitN(r.Notes, matches[0], 2)
	if around[0] == "" {
		newNotes = around[1]
		if around[1] != "" {
			newNotes += "\n\n"
		}
	} else if around[1] == "" {
		newNotes = around[0] + "\n\n"
	} else {
		newNotes = around[0] + matches[1] + around[1] + "\n\n"
	}

	isbn13, err := utils.StandardizeISBN(matches[3])
	if err != nil {
		return false, err
	}

	newNotes += fmt.Sprintf("_%s", isbn13)

	var pages utils.Pages
	var recipeNumber uint64

	if matches[5] != "" {
		pages, err = utils.ParsePages(matches[5])
		if err != nil {
			return false, err
		}

		newNotes += fmt.Sprintf(", p.%s", pages.String())
	}

	if matches[7] != "" && pages != nil {
		recipeNumber, err = strconv.ParseUint(matches[7], 10, 64)
		if err != nil {
			return false, err
		}

		newNotes += fmt.Sprintf(", %s", ordinal(recipeNumber, false))
	}

	newNotes += "_"

	if err := r.SetBook(isbn13, pages, uint(recipeNumber)); err != nil {
		return false, err
	}
	r.Notes = newNotes

	return true, nil
}

func ordinal(n uint64, useWords bool) string {
	if useWords {
		switch n {
		case 1:
			return "first"
		case 2:
			return "second"
		case 3:
			return "third"
		}
	}

	if (n%100)/10 == 1 {
		return fmt.Sprintf("%dth", n)
	}
	switch n % 10 {
	case 1:
		return fmt.Sprintf("%dst", n)
	case 2:
		return fmt.Sprintf("%dnd", n)
	case 3:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}
